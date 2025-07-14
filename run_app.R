#!/usr/bin/env Rscript

# Simple startup script for the Shiny application
library(shiny)

# Run the application
runApp(
  appDir = "main_app.R",
  host = "127.0.0.1",
  port = 3838,
  launch.browser = TRUE
) 