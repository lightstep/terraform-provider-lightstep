package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lightstep/terraform-provider-lightstep/lightstep"
)

func resourceStream() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStreamCreate,
		ReadContext:   resourceStreamRead,
		UpdateContext: resourceStreamUpdate,
		DeleteContext: resourceStreamDelete,
		Importer: &schema.ResourceImporter{
			State: resourceStreamImport,
		},
		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"stream_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"custom_data": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Second),
		},
	}
}

func resourceStreamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*lightstep.Client)
	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		stream, err := client.CreateStream(
			d.Get("project_name").(string),
			d.Get("stream_name").(string),
			d.Get("query").(string),
			d.Get("custom_data").(map[string]interface{}),
		)
		if err != nil {
			// Fix until lock error is resolved
			if strings.Contains(err.Error(), "Internal Server Error") {
				return resource.RetryableError(fmt.Errorf("Expected Creation of stream but not done yet: %s", err))
			} else {
				return resource.NonRetryableError(fmt.Errorf("Error creating stream: %s", err))
			}
		}

		d.SetId(stream.ID)
		if err := resourceStreamRead(ctx, d, m); err != nil {
			if len(err) == 0 {
				return resource.NonRetryableError(fmt.Errorf("Failed to read stream: %v", err))
			}

			return resource.NonRetryableError(fmt.Errorf(err[0].Summary))
		}

		return nil
	}); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to create stream: %v", err))
	}

	return diags
}

func resourceStreamRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*lightstep.Client)
	s, err := client.GetStream(d.Get("project_name").(string), d.Id())
	if err != nil {
		apiErr := err.(lightstep.APIResponseCarrier)
		if apiErr.GetHTTPResponse().StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("Failed to get stream: %v\n", apiErr))
	}

	if err := setResourceDataFromStream(d, *s); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to set stream from API response to terraform state: %v", err))
	}

	return diags
}

func resourceStreamUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*lightstep.Client)

	s := lightstep.Stream{
		Type: "stream",
		ID:   d.Id(),
	}

	if d.HasChange("stream_name") {
		s.Attributes.Name = d.Get("stream_name").(string)
	}

	if _, err := client.UpdateStream(d.Get("project_name").(string), d.Id(), s); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to update stream: %v", err))
	}

	return resourceStreamRead(ctx, d, m)
}

func resourceStreamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*lightstep.Client)
	if err := client.DeleteStream(d.Get("project_name").(string), d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("Failed to detele stream: %v", err))
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}

func resourceStreamImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*lightstep.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Error importing lightstep_stream. Expecting an  ID formed as '<lightstep_project>.<stream_id>'. Got: %v", d.Id())
	}

	project, id := ids[0], ids[1]
	stream, err := client.GetStream(project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Failed to get stream: %v", err)
	}

	d.SetId(id)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Unable to set project_name resource field: %v", err)
	}

	if err := setResourceDataFromStream(d, *stream); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("Failed to set stream from API response to terraform state: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}

func setResourceDataFromStream(d *schema.ResourceData, s lightstep.Stream) error {
	if err := d.Set("stream_name", s.Attributes.Name); err != nil {
		return fmt.Errorf("Unable to set stream_name resource field: %v", err)
	}

	if err := d.Set("custom_data", s.Attributes.CustomData); err != nil {
		return fmt.Errorf("Unable to set custom_data resource field: %v", err)
	}

	if err := d.Set("query", s.Attributes.Query); err != nil {
		return fmt.Errorf("Unable to set query resource field: %v", err)
	}

	return nil
}
