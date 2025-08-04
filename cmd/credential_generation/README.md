# Credential Generation for Microsoft To Do

## How to create OAuth credentials

- register an app in azure [https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps]
- create a Web App with "Accounts in any organizational directory (Any Azure AD directory  - Multitenant) and personal Microsoft accounts (e.g. Skype, Xbox)"
- add redirect url "http://localhost"
- create a secret
  - go to `Manage`
  - select `Certificates & secrets`
  - click `Client secrets`
  - click `New client secret`
  - enter a description and expiration
  - click `Add`
  - copy the value of the created secret (not the Secret ID)
- go back to `Overview` to get the `Application (client) ID` and `Directory (tenant) ID`

Run the following command to set the environment variables in PowerShell:

```powershell
$env:CLIENT_ID="your-client-id"; $env:CLIENT_SECRET="your-client-secret"; $env:TENANT_ID="your-tenant-id"; go run main.go
```

or in bash:

```bash
export CLIENT_ID="your-client-id"; export CLIENT_SECRET="your-client-secret"; export TENANT_ID="your-tenant-id"; go run main.go
```

The code will open a browser window for authentication.
After successful login, it will print the access token to the console.
