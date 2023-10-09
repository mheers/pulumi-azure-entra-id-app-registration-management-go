# pulumi-azure-entra-id-app-registration-management-go

In professional azure environments, it is common to not have access to entra id to create and manage app registrations (e.g. for SSO).

This project expects, that multiple (blanco) app registration already exist. It also expects a separate service principal, which is registered as *Cloud Application Administrator* in the blanco app registrations.

Using the credentials of this management service principle this project can set the redirect URIs of and create secrets in the blanco app registrations.


## Usage

1. Create a new stack:

    ```bash
    pulumi stack init dev
    ```

1. Set credentials:
    
    ```bash
    pulumi config set azure-native:clientId <management-service-principal-client-id>
    pulumi config set azure-native:clientSecret <management-service-principal-client-secret> --secret
    pulumi config set azure-native:tenantId <tenant-id>
    pulumi config set azure-native:location westeurope
    ```

1. Run `pulumi up` to preview and deploy changes:

    ```bash
    pulumi up
    ```

## Caveats

- At every deployment a new secret is created
- Uses [msgraph](github.com/microsoftgraph/msgraph-sdk-go) and does not follow the common pulumi lifecycle
- Only works with service principals, as pulumi tokens can not directly be used to authenticate against msgraph (wrong audience)
