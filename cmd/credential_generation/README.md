# Credential Generation for Microsoft To Do

## How to create OAuth credentials

- register an app in azure [https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps]
- create a Web App with "Accounts in any organizational directory (Any Azure AD directory  - Multitenant) and personal Microsoft accounts (e.g. Skype, Xbox)"
- add "Web platform" with redirect url "<http://localhost>"
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

The code will open a browser window for authentication. After successful login, you will be redirected to http://localhost:7861 and the tool will print the access token to the console and write oauth_credentials.json containing both access_token and refresh_token.

Troubleshooting:
- If oauth_credentials.json has empty tokens:
  - Ensure the redirect URI in Azure AD matches exactly: http://localhost:7861 (including the port).
  - Use TENANT_ID=common if you plan to sign in with a personal Microsoft account; otherwise sign in with an account from the specified tenant.
  - Confirm Microsoft Graph delegated permission "Tasks.ReadWrite" is added and, for organizational tenants, admin consent is granted.
  - The tool requests the offline_access scope to receive a refresh_token. On the first consent, you may need to consent explicitly in the browser.
  - Re-run the generator after fixing the above; you should see a non-empty access token in the console and both access_token and refresh_token in oauth_credentials.json.
