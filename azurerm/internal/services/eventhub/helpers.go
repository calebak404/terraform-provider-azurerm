package eventhub

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
)

// schema
func expandEventHubAuthorizationRuleRights(d *schema.ResourceData) []string {
	rights := make([]string, 0)

	if d.Get("listen").(bool) {
		rights = append(rights, "Listen")
	}

	if d.Get("send").(bool) {
		rights = append(rights, "Send")
	}

	if d.Get("manage").(bool) {
		rights = append(rights, "Manage")
	}

	return rights
}

func flattenEventHubAuthorizationRuleRights(rights []string) (listen, send, manage bool) {
	// zero (initial) value for a bool in go is false

	for _, right := range rights {
		switch right {
		case "Listen":
			listen = true
		case "Send":
			send = true
		case "Manage":
			manage = true
		default:
			log.Printf("[DEBUG] Unknown Authorization Rule Right '%s'", right)
		}
	}

	return listen, send, manage
}

func eventHubAuthorizationRuleSchemaFrom(s map[string]*schema.Schema) map[string]*schema.Schema {
	authSchema := map[string]*schema.Schema{
		"listen": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},

		"manage": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},

		"primary_connection_string": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		"primary_connection_string_alias": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		"primary_key": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		"secondary_connection_string": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		"secondary_connection_string_alias": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		"secondary_key": {
			Type:      schema.TypeString,
			Computed:  true,
			Sensitive: true,
		},

		"send": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
	}
	return azure.MergeSchema(s, authSchema)
}

func eventHubAuthorizationRuleCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, _ interface{}) error {
	listen, hasListen := d.GetOk("listen")
	send, hasSend := d.GetOk("send")
	manage, hasManage := d.GetOk("manage")

	if !hasListen && !hasSend && !hasManage {
		return fmt.Errorf("One of the `listen`, `send` or `manage` properties needs to be set")
	}

	if manage.(bool) && !listen.(bool) && !send.(bool) {
		return fmt.Errorf("if `manage` is set both `listen` and `send` must be set to true too")
	}

	return nil
}
