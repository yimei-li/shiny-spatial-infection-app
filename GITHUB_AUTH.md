# GitHub Authentication Guide

## Option 1: GitHub CLI Authentication (Recommended)

1. Run the authentication command:
   ```bash
   gh auth login
   ```

2. Follow the prompts:
   - Choose "GitHub.com"
   - Choose "HTTPS"
   - Choose "Yes" to authenticate Git operations
   - Choose "Login with a web browser"
   - Copy the one-time code
   - Open the provided URL in your browser
   - Paste the code and authorize

## Option 2: Personal Access Token

1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Give it a name like "Shiny App Deployment"
4. Select scopes: `repo`, `workflow`
5. Copy the token
6. Set environment variable:
   ```bash
   export GH_TOKEN=your_token_here
   ```

## After Authentication

Once authenticated, run:
```bash
gh repo create shiny-spatial-infection-app --public --description "Spatial Cell Infection Simulation Shiny App with Animation" --source=. --remote=origin --push
```

## Manual Alternative

If you prefer to create the repository manually:

1. Go to https://github.com/new
2. Repository name: `shiny-spatial-infection-app`
3. Description: `Spatial Cell Infection Simulation Shiny App with Animation`
4. Make it Public
5. Don't initialize with README, .gitignore, or license
6. Click "Create repository"
7. Then run: `git push -u origin main` 