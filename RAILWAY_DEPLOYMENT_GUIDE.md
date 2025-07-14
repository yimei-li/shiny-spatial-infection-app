# 🚀 Railway Deployment Guide (Step-by-Step)

## Prerequisites
- GitHub account
- Terminal/Command Prompt access
- Your Shiny application files

## Step 1: Prepare Your Project for GitHub

### 1.1 Initialize Git Repository
```bash
# Navigate to your project directory
cd /Users/mael/Desktop/0ABS/research/aim1-spatial-cell/script/add-dip/36_Shinny_add_video

# Initialize git repository
git init

# Add all files to staging
git add .

# Create initial commit
git commit -m "Initial commit: Viral simulation Shiny app"
```

### 1.2 Create GitHub Repository
1. Go to https://github.com
2. Click the **"+"** button in the top right corner
3. Select **"New repository"**
4. Fill in repository details:
   - **Repository name**: `viral-simulation-app`
   - **Description**: `Interactive viral simulation with DIP particles`
   - **Visibility**: Choose **Public** or **Private**
   - **DO NOT** check "Add a README file"
   - **DO NOT** check "Add .gitignore"
   - **DO NOT** check "Choose a license"
5. Click **"Create repository"**

### 1.3 Connect Local Repository to GitHub
```bash
# Add GitHub repository as remote origin
git remote add origin https://github.com/YOUR_USERNAME/viral-simulation-app.git

# Push code to GitHub
git branch -M main
git push -u origin main
```

**Replace `YOUR_USERNAME` with your actual GitHub username**

## Step 2: Deploy to Railway

### 2.1 Create Railway Account
1. Go to https://railway.app
2. Click **"Login"** in the top right corner
3. Choose **"Login with GitHub"**
4. Authorize Railway to access your GitHub account

### 2.2 Create New Project
1. Once logged in, click **"New Project"**
2. Select **"Deploy from GitHub repo"**
3. You'll see a list of your GitHub repositories
4. Find and click on **"viral-simulation-app"** (or whatever you named it)

### 2.3 Configure Deployment
Railway will automatically detect your `Dockerfile.cloud` and start building:

1. **Build Process**: Railway will automatically:
   - Detect the Dockerfile
   - Install dependencies (R, Go, ImageMagick)
   - Build the Docker image
   - This may take 5-10 minutes

2. **Environment Variables**: 
   - Railway automatically sets the `PORT` environment variable
   - No manual configuration needed

### 2.4 Monitor Deployment
1. **Build Logs**: Click on your project to see build progress
2. **Deployments Tab**: Monitor the deployment status
3. **Wait for Success**: Look for "Build completed successfully"

### 2.5 Access Your Application
1. Once deployment is complete, Railway provides a URL
2. Click on the **"View App"** button or the generated URL
3. Your app will be available at: `https://your-app-name.railway.app`

## Step 3: Verify Deployment

### 3.1 Test Application
1. Open the Railway-provided URL
2. Verify the interface loads correctly
3. Test a simulation with default parameters
4. Check if GIF generation works

### 3.2 Check Logs (if issues occur)
1. In Railway dashboard, go to your project
2. Click on **"Deployments"** tab
3. Click on the latest deployment
4. Review **"Build Logs"** and **"Deploy Logs"**

## Step 4: Custom Domain (Optional)

### 4.1 Add Custom Domain
1. In Railway project dashboard
2. Go to **"Settings"** tab
3. Scroll to **"Custom Domain"** section
4. Click **"Add Domain"**
5. Enter your domain name
6. Follow DNS configuration instructions

## Step 5: Update Your Application

### 5.1 Make Changes Locally
```bash
# Edit your files as needed
# For example, modify main_app.R

# Add changes to git
git add .

# Commit changes
git commit -m "Update: Description of your changes"

# Push to GitHub
git push origin main
```

### 5.2 Automatic Redeployment
- Railway automatically redeploys when you push to GitHub
- No manual intervention needed
- Monitor progress in Railway dashboard

## Troubleshooting

### Build Fails
**Problem**: Docker build fails
**Solution**: 
1. Check build logs for error messages
2. Ensure all files are pushed to GitHub
3. Verify Dockerfile.cloud is correct

### Application Won't Start
**Problem**: App builds but doesn't start
**Solution**:
1. Check deploy logs
2. Ensure port configuration is correct
3. Verify R dependencies are installed

### Out of Memory
**Problem**: Build runs out of memory
**Solution**:
1. Upgrade to Railway Pro plan
2. Optimize Docker image size
3. Remove unnecessary files with .dockerignore

### Port Issues
**Problem**: Application not accessible
**Solution**:
1. Use `Dockerfile.cloud` instead of `Dockerfile`
2. Ensure app listens on `0.0.0.0` not `127.0.0.1`
3. Check PORT environment variable usage

## Commands Summary

```bash
# Setup
git init
git add .
git commit -m "Initial commit: Viral simulation Shiny app"
git remote add origin https://github.com/YOUR_USERNAME/viral-simulation-app.git
git branch -M main
git push -u origin main

# Updates
git add .
git commit -m "Update: Your changes description"
git push origin main
```

## Railway Dashboard URLs
- Main Dashboard: https://railway.app/dashboard
- Your Projects: https://railway.app/dashboard/projects
- Account Settings: https://railway.app/account

## Important Notes

1. **Free Tier**: Railway provides $5/month free usage
2. **Build Time**: Initial deployment takes 5-10 minutes
3. **Auto-Deploy**: Pushes to main branch automatically trigger redeployment
4. **HTTPS**: All Railway apps get automatic HTTPS
5. **Logs**: Always check logs if something doesn't work

## Success Checklist

- [ ] GitHub repository created
- [ ] Code pushed to GitHub
- [ ] Railway account created
- [ ] Project deployed on Railway
- [ ] Application accessible via Railway URL
- [ ] Simulation functionality tested
- [ ] GIF generation working

Your viral simulation app is now live on the internet! 🎉 