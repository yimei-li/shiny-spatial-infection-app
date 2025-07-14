# 🚀 Simple Railway Deployment Steps

## Quick Start (5 Steps)

### Step 1: Run Git Setup Script
```bash
./git_setup.sh
```
- Enter your GitHub username when prompted
- Press Enter for default repository name or type a new one
- Confirm the details

### Step 2: Create GitHub Repository
1. Go to https://github.com
2. Click **"+"** → **"New repository"**
3. Repository name: `viral-simulation-app` (or what you chose)
4. Description: `Interactive viral simulation with DIP particles`
5. Choose **Public** or **Private**
6. **DO NOT** check any boxes (README, .gitignore, license)
7. Click **"Create repository"**

### Step 3: Push Code to GitHub
```bash
git push -u origin main
```

### Step 4: Deploy to Railway
1. Go to https://railway.app
2. Click **"Login"** → **"Login with GitHub"**
3. Click **"New Project"** → **"Deploy from GitHub repo"**
4. Select your `viral-simulation-app` repository
5. Wait 5-10 minutes for build to complete

### Step 5: Access Your App
- Railway will provide a URL like: `https://your-app.railway.app`
- Click the URL to access your live application!

## That's it! 🎉

Your viral simulation app is now live on the internet.

## If Something Goes Wrong

1. **Build fails**: Check build logs in Railway dashboard
2. **App won't start**: Make sure you're using `Dockerfile.cloud`
3. **Need help**: Check `RAILWAY_DEPLOYMENT_GUIDE.md` for detailed troubleshooting

## Commands You'll Need

```bash
# Initial setup
./git_setup.sh

# Push to GitHub
git push -u origin main

# Future updates
git add .
git commit -m "Update: describe your changes"
git push origin main
```

## Important URLs

- GitHub: https://github.com
- Railway: https://railway.app
- Your Repository: https://github.com/YOUR_USERNAME/viral-simulation-app
- Railway Dashboard: https://railway.app/dashboard 