package eventhub

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/eventhub/sdk/consumergroups"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/eventhub/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
)

func EventHubConsumerGroupDataSource() *schema.Resource {
	return &schema.Resource{
		Read: EventHubConsumerGroupDataSourceRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.Any(
					validate.ValidateEventHubConsumerName(),
					validation.StringInSlice([]string{"$Default"}, false),
				),
			},

			"namespace_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.ValidateEventHubNamespaceName(),
			},

			"eventhub_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.ValidateEventHubName(),
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),

			"location": azure.SchemaLocationForDataSource(),

			"user_metadata": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func EventHubConsumerGroupDataSourceRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Eventhub.ConsumerGroupClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id := consumergroups.NewConsumergroupID(subscriptionId, d.Get("resource_group_name").(string), d.Get("namespace_name").(string), d.Get("eventhub_name").(string), d.Get("name").(string))

	resp, err := client.Get(ctx, id)
	if err != nil {
		if response.WasNotFound(resp.HttpResponse) {
			return fmt.Errorf("%s was not found", id)
		}
		return fmt.Errorf("retrieving %s: %+v", id, err)
	}

	d.SetId(id.ID())

	d.Set("name", id.Name)
	d.Set("eventhub_name", id.EventhubName)
	d.Set("namespace_name", id.NamespaceName)
	d.Set("resource_group_name", id.ResourceGroup)

	if model := resp.Model; model != nil {
		if props := model.Properties; props != nil {
			d.Set("user_metadata", props.UserMetadata)
		}
	}

	return nil
}
