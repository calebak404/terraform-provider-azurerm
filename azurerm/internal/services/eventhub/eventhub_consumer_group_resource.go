package eventhub

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/sdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/eventhub/sdk/consumergroups"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/eventhub/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

type ConsumerGroupObject struct {
	Name              string `tfschema:"name"`
	NamespaceName     string `tfschema:"namespace_name"`
	EventHubName      string `tfschema:"eventhub_name"`
	ResourceGroupName string `tfschema:"resource_group_name"`
	UserMetadata      string `tfschema:"user_metadata"`
}

var _ sdk.Resource = ConsumerGroupResource{}
var _ sdk.ResourceWithUpdate = ConsumerGroupResource{}

type ConsumerGroupResource struct {
}

func (r ConsumerGroupResource) ResourceType() string {
	return "azurerm_eventhub_consumer_group"
}

func (r ConsumerGroupResource) Arguments() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.ValidateEventHubConsumerName(),
		},

		"namespace_name": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.ValidateEventHubNamespaceName(),
		},

		"eventhub_name": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validate.ValidateEventHubName(),
		},

		"resource_group_name": azure.SchemaResourceGroupName(),

		"user_metadata": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(1, 1024),
		},
	}
}

func (r ConsumerGroupResource) Attributes() map[string]*schema.Schema {
	return map[string]*schema.Schema{}
}

func (r ConsumerGroupResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			metadata.Logger.Info("Decoding state..")
			var state ConsumerGroupObject
			if err := metadata.Decode(&state); err != nil {
				return err
			}

			metadata.Logger.Infof("creating Consumer Group %q..", state.Name)
			client := metadata.Client.Eventhub.ConsumerGroupClient
			subscriptionId := metadata.Client.Account.SubscriptionId

			id := consumergroups.NewConsumergroupID(subscriptionId, state.ResourceGroupName, state.NamespaceName, state.EventHubName, state.Name)
			existing, err := client.Get(ctx, id)
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for the presence of an existing %s: %+v", id, err)
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			parameters := consumergroups.ConsumerGroup{
				Name: utils.String(state.Name),
				Properties: &consumergroups.ConsumerGroupProperties{
					UserMetadata: utils.String(state.UserMetadata),
				},
			}

			if _, err := client.CreateOrUpdate(ctx, id, parameters); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)
			return nil
		},
		Timeout: 30 * time.Minute,
	}
}

func (r ConsumerGroupResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			id, err := consumergroups.ConsumergroupID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			metadata.Logger.Info("Decoding state..")
			var state ConsumerGroupObject
			if err := metadata.Decode(&state); err != nil {
				return err
			}

			metadata.Logger.Infof("updating Consumer Group %q..", state.Name)
			client := metadata.Client.Eventhub.ConsumerGroupClient

			parameters := consumergroups.ConsumerGroup{
				Name: utils.String(id.Name),
				Properties: &consumergroups.ConsumerGroupProperties{
					UserMetadata: utils.String(state.UserMetadata),
				},
			}

			if _, err := client.CreateOrUpdate(ctx, *id, parameters); err != nil {
				return fmt.Errorf("updating %s: %+v", *id, err)
			}

			return nil
		},
		Timeout: 30 * time.Minute,
	}
}

func (r ConsumerGroupResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Eventhub.ConsumerGroupClient
			id, err := consumergroups.ConsumergroupID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			metadata.Logger.Infof("retrieving Consumer Group %q..", id.Name)
			resp, err := client.Get(ctx, *id)
			if err != nil {
				if response.WasNotFound(resp.HttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			state := ConsumerGroupObject{
				Name:              id.Name,
				NamespaceName:     id.NamespaceName,
				EventHubName:      id.EventhubName,
				ResourceGroupName: id.ResourceGroup,
			}

			if model := resp.Model; model != nil && model.Properties != nil {
				state.UserMetadata = utils.NormalizeNilableString(model.Properties.UserMetadata)
			}

			return metadata.Encode(&state)
		},
		Timeout: 5 * time.Minute,
	}
}

func (r ConsumerGroupResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.Eventhub.ConsumerGroupClient
			id, err := consumergroups.ConsumergroupID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			metadata.Logger.Infof("deleting Consumer Group %q..", id.Name)
			if resp, err := client.Delete(ctx, *id); err != nil {
				if !response.WasNotFound(resp.HttpResponse) {
					return fmt.Errorf("deleting %s: %+v", id, err)
				}
			}

			return nil
		},
		Timeout: 30 * time.Minute,
	}
}

func (r ConsumerGroupResource) ModelObject() interface{} {
	return ConsumerGroupObject{}
}

func (r ConsumerGroupResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return validate.EventHubConsumerGroupID
}
