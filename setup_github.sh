#!/bin/bash

echo "🚀 GitHub Setup Helper"
echo "======================"
echo ""

# Check if GitHub CLI is available
if command -v gh &> /dev/null; then
    echo "✅ GitHub CLI found!"
    echo "Setting up GitHub authentication..."
    gh auth login
    echo ""
    echo "Creating repository..."
    gh repo create shiny-spatial-infection-app --public --description "Spatial Cell Infection Simulation Shiny App with Animation" --source=. --remote=origin --push
else
    echo "❌ GitHub CLI not found"
    echo ""
    echo "📋 Manual Steps Required:"
    echo "1. Go to https://github.com/new"
    echo "2. Repository name: shiny-spatial-infection-app"
    echo "3. Description: Spatial Cell Infection Simulation Shiny App with Animation"
    echo "4. Make it Public"
    echo "5. Don't initialize with README, .gitignore, or license"
    echo "6. Click 'Create repository'"
    echo ""
    echo "After creating the repository, run:"
    echo "git push -u origin main"
    echo ""
    echo "If you need authentication, you can:"
    echo "1. Use GitHub CLI: brew install gh && gh auth login"
    echo "2. Or create a Personal Access Token at https://github.com/settings/tokens"
fi 