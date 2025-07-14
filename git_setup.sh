#!/bin/bash

echo "🚀 GitHub Setup for Railway Deployment"
echo "======================================"
echo ""

# Get GitHub username
read -p "Enter your GitHub username: " github_username

# Get repository name (with default)
read -p "Enter repository name [viral-simulation-app]: " repo_name
repo_name=${repo_name:-viral-simulation-app}

echo ""
echo "📋 Setup Summary:"
echo "GitHub Username: $github_username"
echo "Repository Name: $repo_name"
echo "Repository URL: https://github.com/$github_username/$repo_name.git"
echo ""

read -p "Is this correct? (y/n): " confirm

if [[ $confirm != "y" && $confirm != "Y" ]]; then
    echo "❌ Setup cancelled. Please run the script again."
    exit 1
fi

echo ""
echo "🔧 Setting up Git repository..."

# Check if git is initialized
if [ ! -d ".git" ]; then
    echo "Initializing git repository..."
    git init
else
    echo "Git repository already exists."
fi

# Add all files
echo "Adding files to git..."
git add .

# Create initial commit
echo "Creating initial commit..."
git commit -m "Initial commit: Viral simulation Shiny app"

# Add remote origin
echo "Adding GitHub remote..."
git remote remove origin 2>/dev/null
git remote add origin https://github.com/$github_username/$repo_name.git

# Set main branch
git branch -M main

echo ""
echo "✅ Git setup complete!"
echo ""
echo "📝 Next Steps:"
echo "1. Go to https://github.com and create a new repository:"
echo "   - Repository name: $repo_name"
echo "   - Description: Interactive viral simulation with DIP particles"
echo "   - Make it Public or Private"
echo "   - DO NOT add README, .gitignore, or license"
echo ""
echo "2. After creating the repository, run:"
echo "   git push -u origin main"
echo ""
echo "3. Then follow the Railway deployment guide in RAILWAY_DEPLOYMENT_GUIDE.md"
echo ""
echo "🔗 GitHub Repository URL: https://github.com/$github_username/$repo_name" 