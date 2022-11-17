package lightstep

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/lightstep/terraform-provider-lightstep/client"
	"net/http"
	"strings"
)

// resourceNotebook creates a resource for a Lightstep Notebook
func resourceNotebook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNotebookCreate,
		ReadContext:   resourceNotebookRead,
		DeleteContext: resourceNotebookDelete,
		UpdateContext: resourceNotebookUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNotebookImport,
		},
		DeprecationMessage: "resource_notebook is no longer supported. Please migrate to resource_metric_notebook with span queries.",
		Schema:             getNotebookSchema(),
	}
}

func resourceNotebookCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*client.Client)

	projectName := d.Get("project_name").(string)
	notebookName := d.Get("notebook_name").(string)
	notebookDescription := d.Get("notebook_description").(string)
	var notebookEntries []client.NotebookEntry // TODO

	notebook := client.Notebook{
		Attributes: client.NotebookAttributes{
			Name:        notebookName,
			Description: notebookDescription,
			Entries:     notebookEntries,
		},
	}

	created, err := c.CreateNotebook(ctx, projectName, notebook)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create notebook for [project: %v; notebook: %v]: %v", projectName, notebookName, err))
	}

	d.SetId(created.ID)
	return resourceNotebookRead(ctx, d, m)
}

func resourceNotebookRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)

	projectName := d.Get("project_name").(string)
	notebookID := d.Id()

	notebook, err := c.GetNotebook(ctx, projectName, notebookID)
	if err != nil {
		apiErr := err.(client.APIResponseCarrier)
		if apiErr.GetHTTPResponse().StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("failed to get notebook for [project: %v; resource_id: %v]: %v", projectName, notebookID, apiErr))
	}

	if err := setResourceDataFromNotebook(d, *notebook); err != nil {
		return diag.FromErr(fmt.Errorf("failed to set notebook response from API to terraform state for [project: %v; resource_id: %v]: %v", projectName, notebookID, err))
	}

	return diags
}

func resourceNotebookUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	projectName := d.Get("project_name").(string)
	notebookID := d.Id()
	notebookName := d.Get("notebook_name").(string)
	notebookDescription := d.Get("notebook_description").(string)
	var notebookEntries []client.NotebookEntry // TODO

	notebook := client.Notebook{
		ID: notebookID,
		Attributes: client.NotebookAttributes{
			Name:        notebookName,
			Description: notebookDescription,
			Entries:     notebookEntries,
		},
	}
	if _, err := c.UpdateNotebook(ctx, projectName, notebookID, notebook); err != nil {
		return diag.FromErr(fmt.Errorf("failed to update condition for [project: %v; notebook_name: %v, resource_id: %v]: %v", projectName, notebookName, notebookID, err))
	}

	return diags
}

func resourceNotebookDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := m.(*client.Client)
	projectName := d.Get("project_name").(string)
	notebookID := d.Id()

	if err := c.DeleteNotebook(ctx, projectName, notebookID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete notebook for [project: %v; resource_id: %v]: %v", projectName, notebookID, err))
	}

	d.SetId("")
	return diags
}

func resourceNotebookImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	c := m.(*client.Client)

	notebookID := d.Id()
	ids := strings.Split(notebookID, ".")
	if len(ids) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("error importing lightstep_notebook. Expecting an  ID formed as '<lightstep_project>.<lightstep_notebookID>' (provided: %v)", notebookID)
	}
	project, id := ids[0], ids[1]

	notebook, err := c.GetNotebook(ctx, project, id)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(id)
	if err := d.Set("project_name", project); err != nil {
		return []*schema.ResourceData{}, err
	}

	if err := setResourceDataFromNotebook(d, *notebook); err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("failed to set notebook from API response to terraform state: %v", err)
	}

	return []*schema.ResourceData{d}, nil
}

func setResourceDataFromNotebook(d *schema.ResourceData, notebook client.Notebook) error {
	if err := d.Set("notebook_name", notebook.Attributes.Name); err != nil {
		return fmt.Errorf("unable to set notebook_name resource field: %v", err)
	}

	if err := d.Set("notebook_description", notebook.Attributes.Name); err != nil {
		return fmt.Errorf("unable to set notebook_description resource field: %v", err)
	}

	var entries []interface{}
	for _, e := range notebook.Attributes.Entries {
		entry := map[string]interface{}{}

		entry["id"] = e.ID
		entry["type"] = e.Type
		entry["start_micros"] = e.StartMicros
		entry["end_micros"] = e.EndMicros
		entry["rank"] = e.ID

		if e.TextBlock != nil {
			textBlock := map[string]interface{}{}
			textBlock["text"] = e.TextBlock.Text
			entry["text_block"] = textBlock

		} else if e.Chart != nil {
			chart := map[string]interface{}{}

			chart["id"] = e.Chart.ID
			chart["title"] = e.Chart.Title
			chart["subtitle"] = e.Chart.Subtitle
			chart["chart_type"] = e.Chart.ChartType

			qStrings := make([]string, 0, len(e.Chart.QueryStrings))
			for _, q := range e.Chart.QueryStrings {
				qStrings = append(qStrings, q)
			}
			chart["query_strings"] = qStrings

			yAxis := map[string]interface{}{}
			if e.Chart.YAxis != nil {
				yAxis["max"] = e.Chart.YAxis.Max
				yAxis["min"] = e.Chart.YAxis.Min
				chart["y_axis"] = []map[string]interface{}{yAxis}
			}

			entry["chart"] = chart
		}
	}

	if err := d.Set("notebook_entries", entries); err != nil {
		return fmt.Errorf("unable to set notebook_entries resource field: %v", err)
	}

	return nil
}

func getNotebookSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"notebook_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"notebook_description": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"notebook_entries": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getNotebookEntrySchema(),
			},
		},
	}
}

func getNotebookEntrySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"chart", "text_block"}, false),
		},
		"start_micros": {
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Required:     true,
		},
		"end_micros": {
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Required:     true,
		},
		"rank": {
			Type:         schema.TypeInt,
			ValidateFunc: validation.IntAtLeast(0),
			Required:     true,
		},
		"text_block": {
			Optional: true,
			Type:     schema.TypeList,
			Elem: &schema.Resource{
				Schema: getNotebookTextBlockSchema(),
			},
		},
		"chart": {
			Optional: true,
			Type:     schema.TypeList,
			Elem: &schema.Resource{
				Schema: getNotebookChartSchema(),
			},
		},
	}
}

func getNotebookTextBlockSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"text": {
			Type:     schema.TypeString,
			Required: true,
		},
	}
}

func getNotebookChartSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"subtitle": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"chart_type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"timeseries"}, false),
		},
		"query_strings": {
			Type:     schema.TypeList,
			Required: true,
			Elem:     schema.TypeString,
		},
		"y_axis": {
			Type:       schema.TypeList,
			MaxItems:   1,
			Deprecated: "The y_axis field is no longer used",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"min": {
						Type:     schema.TypeFloat,
						Required: true,
					},
					"max": {
						Type:     schema.TypeFloat,
						Required: true,
					},
				},
			},
			Optional: true,
		},
	}
}
