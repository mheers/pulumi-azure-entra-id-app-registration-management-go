package main

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	graphapplications "github.com/microsoftgraph/msgraph-sdk-go/applications"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/pulumi/pulumi-azuread/sdk/v5/go/azuread"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		err := run(ctx)
		if err != nil {
			return err
		}
		return nil
	})
}

func run(ctx *pulumi.Context) error {
	appRegistrationName := "demo"
	endpoints := []string{"https://marcelheers:8080"}

	app, err := azuread.LookupApplication(ctx, &azuread.LookupApplicationArgs{
		DisplayName: pulumi.StringRef(appRegistrationName),
	}, nil)
	if err != nil {
		return err
	}
	ctx.Export("applicationId", pulumi.String(app.ApplicationId))
	ctx.Export("applicationObjectId", pulumi.String(app.ObjectId))

	graphClient, err := graphClient(ctx)
	if err != nil {
		return err
	}

	if !ctx.DryRun() {
		secret, err := createSecret(graphClient, app.ObjectId)
		if err != nil {
			return err
		}
		ctx.Export("secret", pulumi.String(secret))
	}

	if !ctx.DryRun() {
		err = setRedirectURIs(graphClient, app.ApplicationId, endpoints)
		if err != nil {
			return err
		}

		sp, err := azuread.LookupServicePrincipal(ctx, &azuread.LookupServicePrincipalArgs{
			ApplicationId: pulumi.StringRef(app.ApplicationId),
		}, nil)
		if err != nil {
			return err
		}
		ctx.Export("redirect URIs of sp", pulumi.ToStringArray(sp.RedirectUris))
	}

	return nil
}

func creds(ctx *pulumi.Context) (*azidentity.ClientSecretCredential, error) {
	tenantId, exists := ctx.GetConfig("azure-native:tenantId")
	if !exists {
		return nil, fmt.Errorf("azure-native:tenantId not found")
	}

	clientId, exists := ctx.GetConfig("azure-native:clientId")
	if !exists {
		return nil, fmt.Errorf("azure-native:clientId not found")
	}

	clientSecret, exists := ctx.GetConfig("azure-native:clientSecret")
	if !exists {
		return nil, fmt.Errorf("azure-native:clientSecret not found")
	}

	return azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, nil)
}

func graphClient(ctx *pulumi.Context) (*msgraphsdkgo.GraphServiceClient, error) {
	creds, err := creds(ctx)
	if err != nil {
		return nil, err
	}

	client, err := msgraphsdkgo.NewGraphServiceClientWithCredentials(creds, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func setRedirectURIs(client *msgraphsdkgo.GraphServiceClient, appId string, uris []string) error {
	requestBody := models.NewApplication()
	webApp := models.NewWebApplication()
	webApp.SetRedirectUris(uris)
	requestBody.SetWeb(webApp)

	_, err := client.ApplicationsWithAppId(&appId).Patch(context.Background(), requestBody, nil)
	if err != nil {
		return err
	}

	return nil
}

func createSecret(client *msgraphsdkgo.GraphServiceClient, objectId string) (string, error) {
	requestBody := graphapplications.NewItemAddPasswordPostRequestBody()
	passwordCredential := models.NewPasswordCredential()
	displayName := "created by pulumi"
	passwordCredential.SetDisplayName(&displayName)
	requestBody.SetPasswordCredential(passwordCredential)

	result, err := client.Applications().ByApplicationId(objectId).AddPassword().Post(context.Background(), requestBody, nil)
	if err != nil {
		return "", err
	}

	password := result.GetSecretText()

	return *password, nil
}
