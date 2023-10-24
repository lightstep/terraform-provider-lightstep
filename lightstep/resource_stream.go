package lightstep

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lightstep/terraform-provider-lightstep/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceStream() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceStreamCreate,
		ReadContext:   resourceStreamRead,
		UpdateContext: resourceStreamUpdate,
		DeleteContext: resourceStreamDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceStreamImport,
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
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Second),
		},
	}
}

// resourceStreamCreate creates a new stream if needed (if one doesn't exist for this query predicate)
// or imports the existing stream resource into the terraform state
func resourceStreamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	if err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		origQuery := d.Get("query").(string)
		stream, err := c.CreateStream(
			ctx,
			d.Get("project_name").(string),
			d.Get("stream_name").(string),
			d.Get("query").(string),
			d.Get("custom_data").([]interface{}),
		)
		if err != nil {
			// Fix until lock error is resolved
			if strings.Contains(err.Error(), "Internal Server Error") {
				return resource.RetryableError(fmt.Errorf("expected Creation of stream but not done yet: %s", err))
			} else {
				return resource.NonRetryableError(fmt.Errorf("error creating stream: %s", err))
			}
		}

		d.SetId(stream.ID)
		if err := resourceStreamRead(ctx, d, m); err != nil {
			if len(err) == 0 {
				return resource.NonRetryableError(fmt.Errorf("failed to read stream: %v", err))
			}

			return resource.NonRetryableError(fmt.Errorf(err[0].Summary))
		}
		// workaround: if read succeeds, persist the *client-side* query expression to avoid backend normalization issue
		d.Set("query", origQuery)
		return nil
	}); err != nil {
		return diag.FromErr(fmt.Errorf("failed to create stream: %v", err))
	}

	return diags
}

func resourceStreamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	s, err := c.GetStream(ctx, d.Get("project_name").(string), d.Id())
	if err != nil {
		apiErr, isApiErr := err.(client.APIResponseCarrier)

		if isApiErr {
			resp := apiErr.GetHTTPResponse()
			if resp == nil {
				return diag.FromErr(fmt.Errorf("failed to get stream response: %v", err))
			}

			if resp.StatusCode == http.StatusNotFound {
				d.SetId("")
				return diags
			}
		}

		return diag.FromErr(fmt.Errorf("failed to get stream: %v", err))
	}

	if err := setResourceDataFromStream(d, *s); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set stream from API response to terraform state: %v", err))
	}

	return diags
}

func resourceStreamUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	s := client.Stream{
		Type: "stream",
		ID:   d.Id(),
	}

	s.Attributes.Name = d.Get("stream_name").(string)
	s.Attributes.CustomData = client.CustomDataConvert(d.Get("custom_data").([]interface{}))

	if _, err := c.UpdateStream(ctx, d.Get("project_name").(string), d.Id(), s); err != nil {
		return diag.FromErr(fmt.Errorf("failed to update stream: %v", err))
	}

	return resourceStreamRead(ctx, d, m)
}

func resourceStreamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	if err := c.DeleteStream(ctx, d.Get("project_name").(string), d.Id()); err != nil {
		if apiClientError, ok := err.(client.APIResponseCarrier); ok && apiClientError.GetStatusCode() == http.StatusConflict {
			// Lightstep didn't delete the stream itself because there are other resources
			// (usually alerts) that depend on it. However, that relationship is often implicit,
			// and thus opaque to the terraform user, so we do our best to satisfy the user's
			// intent by acting as though the delete operation succeed. This will cause the stream
			// resource to be removed from the terraform state, leaving the stream intact in Lightstep
			// but no longer managed by terraform.
		} else {
			return diag.FromErr(fmt.Errorf("failed to delete stream: %v", err))
		}
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")
	return diags
}

func resourceStreamImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	ids := strings.Split(d.Id(), ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("error importing lightstep_stream. Expecting an  ID formed as '<lightstep_project>.<stream_id>'. Got: %v", d.Id())
	}

	project, id := ids[0], ids[1]
	stream, err := c.GetStream(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to get stream: %v", err)
	}

	d.SetId(id)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to set project_name resource field: %v", err)
	}

	if err := setResourceDataFromStream(d, *stream); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to set stream from API response to terraform state: %v", err)
	}
	d.Set("query", stream.Attributes.Query)

	return []*schema.ResourceData{d}, nil
}

func setResourceDataFromStream(d *schema.ResourceData, s client.Stream) error {
	if err := d.Set("stream_name", s.Attributes.Name); err != nil {
		return fmt.Errorf("unable to set stream_name resource field: %v", err)
	}

	// Convert custom_data to list
	customData := []map[string]string{}

	// This is what Lightstep sends
	//"custom_data": {
	//	"object1": {
	//		"url": "http://",
	//		"key": "value"
	//	},
	//	"object2": {
	//		"key": "value"
	//	}
	//},

	// This is what terraform expects
	//	custom_data = [
	//	  {
	//      // This name field is special and becomes the key
	//	  	"name": "object1"
	//  	"url" = "https://lightstep.atlassian.net/l/c/M7b0rBsj",
	//      "key" = "value",
	//    },
	//  ]
	// Hack until https://lightstep.atlassian.net/browse/LS-26494 is fixed.
	for name, data := range s.Attributes.CustomDataGet {
		d := make(map[string]string)

		d["name"] = name
		for k, v := range data {
			// k is "object1"
			// v is map of key,values
			d[k] = v
		}

		customData = append(customData, d)
	}

	if err := d.Set("custom_data", customData); err != nil {
		return fmt.Errorf("unable to set custom_data resource field: %v", err)
	}

	// don't set query here to avoid backend normalization issue

	return nil
}
