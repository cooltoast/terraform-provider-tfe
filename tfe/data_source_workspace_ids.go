package tfe

import (
	"fmt"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceTFEWorkspaceIDs() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "\"ids\": [DEPRECATED] Use full_names instead. The ids attribute will be removed in the future. See the CHANGELOG to learn more: https://github.com/hashicorp/terraform-provider-tfe/blob/v0.18.0/CHANGELOG.md",
		Read:               dataSourceTFEWorkspaceIDsRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},

			"organization": {
				Type:     schema.TypeString,
				Required: true,
			},

			"ids": {
				Type:       schema.TypeMap,
				Computed:   true,
				Deprecated: "Use full_names instead. The ids attribute will be removed in the future.",
			},

			"external_ids": {
				Type:     schema.TypeMap,
				Computed: true,
			},

			"full_names": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceTFEWorkspaceIDsRead(d *schema.ResourceData, meta interface{}) error {
	tfeClient := meta.(*tfe.Client)

	// Get the organization.
	organization := d.Get("organization").(string)

	// Create a map with all the names we are looking for.
	var id string
	names := make(map[string]bool)
	for _, name := range d.Get("names").([]interface{}) {
		id += name.(string)
		names[name.(string)] = true
	}

	// Create two maps to hold the results.
	ids := make(map[string]string, len(names))
	externalIDs := make(map[string]string, len(names))

	options := tfe.WorkspaceListOptions{}
	for {
		wl, err := tfeClient.Workspaces.List(ctx, organization, options)
		if err != nil {
			return fmt.Errorf("Error retrieving workspaces: %v", err)
		}

		for _, w := range wl.Items {
			if names["*"] || names[w.Name] {
				ids[w.Name] = organization + "/" + w.Name
				externalIDs[w.Name] = w.ID
			}
		}

		// Exit the loop when we've seen all pages.
		if wl.CurrentPage >= wl.TotalPages {
			break
		}

		// Update the page number to get the next page.
		options.PageNumber = wl.NextPage
	}

	d.Set("ids", ids)
	d.Set("external_ids", externalIDs)
	d.Set("full_names", ids)
	d.SetId(fmt.Sprintf("%s/%d", organization, schema.HashString(id)))

	return nil
}
