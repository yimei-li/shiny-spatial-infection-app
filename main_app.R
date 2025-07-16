# Set CRAN mirror
options(repos = c(CRAN = "https://cloud.r-project.org/"))

# Install and load required packages
if (!requireNamespace("shiny", quietly = TRUE)) {
  install.packages("shiny")
}
if (!requireNamespace("shinyjs", quietly = TRUE)) {
  install.packages("shinyjs")
}
library(shiny)
library(shinyjs)

# Source GIF generator functions
source("gif_generator.R")

# Define User Interface (UI)
ui <- fluidPage(
  # Google Fonts & Custom CSS
  tags$head(
    tags$link(rel = "stylesheet", href = "https://fonts.googleapis.com/css?family=Roboto:400,700&display=swap"),
    tags$style(HTML("
      body {
        background: #fafafa;
        font-family: 'Roboto', 'Helvetica Neue', Arial, sans-serif;
        color: #333333;
        margin: 0;
        padding: 0 30px;
      }
      .container-fluid {
        padding-top: 0px;
        padding-bottom: 36px;
        padding-left: 20px;
        padding-right: 20px;
      }
      .row {
        display: flex;
        flex-wrap: wrap;
        margin-right: 0px;
        margin-left: 0px;
        gap: 20px;
        align-items: flex-start;
      }
      .left-panel, .right-panel {
        background: #ffffff;
        color: #333333;
        border-radius: 16px;
        box-shadow: 0 4px 32px rgba(0,0,0,0.08);
        border: 1px solid #e0e0e0;
        margin-bottom: 24px;
        padding: 20px;
        position: relative;
        padding-left: 25px;
        padding-right: 25px;
      }
      .left-panel {
        flex: 0 0 380px;
        min-width: 350px;
        max-width: 480px;
        margin-left: 10px;
      }
      .right-panel {
        flex: 1;
        min-width: 500px;
        margin-right: 30px;
      }
      .result-area {
        background: #f8f9fa;
        border: 2px dashed #ff8c00;
        color: #555555;
        min-height: 320px;
        border-radius: 16px;
        font-size: 1.1em;
        margin-top: 20px;
        margin-bottom: 20px;
        box-shadow: 0 2px 20px rgba(255,140,0,0.07);
      }
      #run_simulation {
        font-size: 24px;
        padding: 18px 36px;
        font-weight: bold;
        width: 100%;
        margin: 32px 0 24px 0;
        background: linear-gradient(90deg, #ff8c00 0%, #ff6b35 100%);
        color: #fff;
        border: none;
        border-radius: 12px;
        box-shadow: 0 2px 12px rgba(255,140,0,0.3);
        transition: background 0.3s, box-shadow 0.3s;
        letter-spacing: 2px;
      }
      #run_simulation:hover {
        background: linear-gradient(90deg, #ff6b35 0%, #ff8c00 100%);
        box-shadow: 0 4px 24px rgba(255,107,53,0.35);
      }
      h2, h3, h4, h5 {
        color: #ff8c00;
        font-weight: 700;
        letter-spacing: 1px;
      }
      .form-control, .selectize-input {
        background: #ffffff !important;
        color: #333333 !important;
        border: 2px solid #ff8c00 !important;
        border-radius: 8px !important;
        margin-bottom: 12px;
        font-size: 1.2em !important;
        box-shadow: none !important;
        font-weight: 500 !important;
      }
      .form-control:focus, .selectize-input.focus {
        border-color: #ff6b35 !important;
        box-shadow: 0 0 0 0.2rem rgba(255,140,0,0.25) !important;
      }
      .selectize-dropdown {
        background: #ffffff !important;
        color: #333333 !important;
        font-size: 1.2em !important;
        border: 1px solid #ff8c00 !important;
      }
      label {
        color: #333333;
        font-weight: 600;
        font-size: 1.1em !important;
        margin-bottom: 8px;
        line-height: 1.3;
      }
      .form-control:disabled {
        background-color: #f5f5f5 !important;
        color: #999999 !important;
        cursor: not-allowed !important;
        opacity: 0.7 !important;
      }
      .form-control:disabled + label {
        color: #999999 !important;
      }
      #progress_status {
        position: fixed;
        bottom: 24px;
        right: 24px;
        background: rgba(255, 140, 0, 0.95);
        color: #fff;
        padding: 14px 28px;
        border-radius: 8px;
        font-size: 16px;
        z-index: 1000;
        display: none;
        box-shadow: 0 2px 16px rgba(255,140,0,0.3);
        letter-spacing: 1px;
      }
      hr {
        border-top: 2px solid #ff8c00 !important;
      }
      /* Scrollbar styling */
      ::-webkit-scrollbar {
        width: 8px;
        background: #f0f0f0;
      }
      ::-webkit-scrollbar-thumb {
        background: #ff8c00;
        border-radius: 4px;
      }
      /* Responsive for medium screens */
      @media (max-width: 1200px) {
        .left-panel {
          margin-left: 20px;
        }
        .right-panel {
          margin-right: 20px;
          min-width: 400px;
        }
        body {
          padding: 0 10px;
        }
        .container-fluid {
          padding-left: 10px;
          padding-right: 10px;
        }
      }
      
      /* Responsive for small screens */
      @media (max-width: 900px) {
        .row {
          flex-direction: column;
          gap: 15px;
        }
        .left-panel, .right-panel {
          flex: none;
          width: calc(100% - 40px);
          margin-left: 20px;
          margin-right: 20px;
          min-width: unset;
          max-width: unset;
        }
        .left-panel {
          order: 1;
          margin-bottom: 20px;
        }
        .right-panel {
          order: 2;
          margin-top: 0px;
        }
        .result-area {
          min-height: 250px;
        }
        body {
          padding: 0 5px;
        }
        .container-fluid {
          padding-left: 5px;
          padding-right: 5px;
        }
      }
      
      /* Responsive for mobile screens */
      @media (max-width: 600px) {
        .left-panel, .right-panel {
          width: calc(100% - 20px);
          margin-left: 10px;
          margin-right: 10px;
          padding: 15px;
        }
        .title-container {
          padding: 20px 10px 15px 10px;
        }
        .result-area {
          min-height: 200px;
        }
        #run_simulation {
          font-size: 20px;
          padding: 15px 25px;
        }
      }
      .title-container {
        background-color: #000000;
        color: #ffffff;
        padding: 50px 20px 30px 20px;
        margin: 0 0 30px 0;
        font-family: 'Roboto', sans-serif;
        text-align: center;
        width: 100%;
        position: relative;
        top: 0px;
        margin-top: 0px;
      }
      .copyright-footer {
        background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
        border-top: 2px solid #ff8c00;
        text-align: center;
        margin-top: 50px;
        padding: 25px 20px;
        color: #495057;
        font-size: 0.9em;
        box-shadow: 0 -2px 10px rgba(0,0,0,0.05);
      }
      .copyright-footer strong {
        color: #ff8c00;
        font-weight: 700;
      }
    "))
  ),
  useShinyjs(),

  div(class = "title-container",
          h1(style="color:#ffffff; letter-spacing:3px; font-weight:900; font-size:3.5em; font-family: 'Roboto', sans-serif; margin: 0;",
         icon("flask", class = "fa-lg", style="color:#42a5f5; margin-right:15px;"),
         "Spatio-temporal Virus-DIP-IFN Simulation Platform"
      )
  ),
  br(),

  # Two-column layout with 1:3 ratio
  fluidRow(
    # Left column - Parameters (25% width)
    column(3, class = "left-panel",
      h2("Parameter Settings"),
      hr(style="border-top:2px solid #42a5f5;"),
      h3("Simulation Parameters"),
      h4("Simulation Configuration"),
      selectInput("particleSpreadOption", "Viral Particle Spread Mechanism",
                  choices = c("celltocell", "global", "local"),
                  selected = "celltocell"),
      selectInput("ifnSpreadOption", "IFN Signal Spread Pattern",
                  choices = c("global", "local", "celltocell"),
                  selected = "global"),
      checkboxInput("dipOption", "Enable DIP Particles", value = TRUE),
      numericInput("dipSynthesisAdvantage", "DIP Synthesis Speed Advantage", value = 4, min = 1, max = 20),
      tags$p(style = "color: #666; font-size: 0.9em; margin-top: -8px; margin-bottom: 12px;", 
             "Fold advantage in synthesis speed of DIPs compared to full-length virions"),
      h4("Infection & Cell Dynamics"),
      numericInput("tau", "IFN Response Delay Time", value = 95, min = 1, max = 200),
      tags$p(style = "color: #666; font-size: 0.9em; margin-top: -8px; margin-bottom: 12px;", 
             "Mean delay for cells to enter antiviral state after IFN stimulation"),
      numericInput("burstSizeV", "Virion Burst Size per Cell", value = 50, min = 1, max = 1000),
      numericInput("burstSizeD", "DIP Burst Size per Cell", value = 100, min = 1, max = 5000),
      numericInput("meanLysisTime", "Mean Cell Lysis Time", value = 12.0, min = 0.1, max = 100, step = 0.1),
      textInput("kJumpR", "Random Jump Ratio", value = "NaN"),
      tags$p(style = "color: #666; font-size: 0.9em; margin-top: -8px; margin-bottom: 12px;", 
             "Fraction of particles spreading randomly; The rest spreads cell-to-cell"),
      numericInput("ifnBothFold", "IFN Fold Increase for Co-infection", value = 1.0, min = 0.1, max = 10, step = 0.1),
      numericInput("rho", "Base Infection Probability", value = 0.026, min = 0.001, max = 1, step = 0.001),
      tags$p(style = "color: #666; font-size: 0.9em; margin-top: -8px; margin-bottom: 12px;", 
             "Probability of successful infection per virus-cell contact event"),
      h4("Particle & Signal Decay"),
      numericInput("virion_half_life", "Virion Decay Half-life", value = 0.0, min = 0.0, max = 100, step = 0.1),
      numericInput("dip_half_life", "DIP Decay Half-life", value = 0.0, min = 0.0, max = 100, step = 0.1),
      numericInput("ifn_half_life", "IFN Signal Decay Half-life", value = 0.0, min = 0.0, max = 100, step = 0.1),
      h4("Visualization Options"),
      selectInput("videotype", "Output Visualization Type",
                  choices = c("states", "IFNconcentration"),
                  selected = "states")
    ),

    # Right column - Run & Results (75% width)
    column(9, class = "right-panel",
      h2("Results Display"),
      actionButton("run_simulation", "RUN SIMULATION", icon = icon("play"), class = "btn-primary btn-lg"),
      br(), br(),
      div(id = "progress_display", 
          style = "background: #f8f9fa; border: 1px solid #dee2e6; border-radius: 8px; padding: 15px; margin-bottom: 20px; min-height: 60px; display: none;",
          h3("Simulation Progress", style = "margin-top: 0; color: #ff8c00; margin-left: 40px;"),
          div(id = "progress_text", "")
      ),
      div(class = "result-area",
        h3("Simulation Results", style = "margin-left: 40px;"),
        br(),
        div(id = "simulation_results",
          p("Results will appear here after running the simulation.", style = "margin-left: 40px;")
        )
      ),
      div(id = "progress_status", "")
    )
  ),
  
  # Copyright Footer
  div(class = "copyright-footer",
      p(style = "margin: 0; line-height: 1.8;",
        "© 2024 Yimei Li. All rights reserved.", br(),
        "Princeton University, Grenfell Lab / teVelthuis Lab / Levin Lab", br(),
        strong("For non-commercial use only. Commercial use is strictly prohibited."), br(),
        em("Spatio-temporal Virus-DIP-IFN Simulation Platform")
      )
  )
)

# Function to generate results display HTML
generate_results_display <- function(input, option_value, go_output) {
  # Look for output folder (Go script creates folders with complex names)
  output_folders <- list.dirs(".", recursive = FALSE, full.names = TRUE)
  # Filter for folders that contain typical Go script output files
  output_folders <- output_folders[sapply(output_folders, function(folder) {
    csv_exists <- file.exists(file.path(folder, "simulation_output.csv"))
    png_exists <- file.exists(file.path(folder, "selected_frames_combined.png"))
    return(csv_exists || png_exists)
  })]
  
  if (length(output_folders) > 0) {
    # Get the most recent output folder
    latest_folder <- output_folders[which.max(file.info(output_folders)$mtime)]
    
    # Check for output files
    csv_file <- file.path(latest_folder, "simulation_output.csv")
    png_file <- file.path(latest_folder, "selected_frames_combined.png")
    mp4_file <- file.path(latest_folder, "video.mp4")
    
    # Start with results status
    html_content <- paste0(
      "<h4 style='color:#28a745; margin-left: 40px;'>Simulation Complete!</h4>",
      "<div style='text-align: left; background: #ffffff; color:#333333; padding: 15px; border-radius: 8px; margin: 10px 0; border-left:4px solid #28a745; border: 1px solid #e0e0e0;'>",
      "<h5 style='color:#28a745; margin-left: 40px;'>Results:</h5>",
      "<p style='margin-left: 40px;'><strong>Status:</strong> ✅ Executed successfully</p>"
    )
    
    # Try to generate GIF animation from combined frames, but only if PNG exists
    gif_path <- NULL
    if (file.exists(png_file)) {
      tryCatch({
        gif_path <- create_gif_from_simulation_folder(latest_folder)
      }, error = function(e) {
        cat("GIF generation failed:", e$message, "\n")
        gif_path <- NULL
      })
    }
    
    if (!is.null(gif_path) && file.exists(gif_path)) {
      # GIF was created successfully
      gif_name <- basename(gif_path)
      
      html_content <- paste0(html_content,
        "<p style='margin-left: 40px;'><strong>Simulation Animation:</strong></p>",
        "<div style='text-align: center; margin: 20px 0; background: #f8f9fa; padding: 20px; border-radius: 8px; border: 1px solid #ddd;'>",
        "<img src='www/", gif_name, "' alt='Simulation Animation' style='width: 80%; min-width: 400px; max-width: 600px; height: auto; border: 2px solid #007bff; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);'>",
        "<br/><br/>",
        "<p style='color: #333; font-size: 24px; line-height: 1.5; margin: 15px 0 10px 0; font-weight: 400;'><strong>Simulation Description:</strong> The above lines show cell state percentages over time. The plaque simulation below shows how virus and DIP particles spread from a single cell infection center, presenting the plaque morphology over 500 hours. Colors represent different cell states:</p>",
        "<div style='background: #f8f9fa; padding: 15px; border-radius: 8px; margin: 10px 0;'>",
        "<ul style='color: #333; font-size: 18px; line-height: 1.4; margin: 0; padding-left: 0; list-style: none;'>",
        "<li style='margin-bottom: 4px; padding: 3px 0;'><span style='color: #000000; font-weight: bold; font-size: 20px;'>●</span> <strong>Black</strong> - Uninfected/susceptible cells</li>",
        "<li style='margin-bottom: 4px; padding: 3px 0;'><span style='color: #808080; font-weight: bold; font-size: 20px;'>●</span> <strong>Grey</strong> - Dead cells</li>",
        "<li style='margin-bottom: 4px; padding: 3px 0;'><span style='color: #FFD700; font-weight: bold; font-size: 20px;'>●</span> <strong>Yellow</strong> - Cells infected by both DIPs and virions</li>",
        "<li style='margin-bottom: 4px; padding: 3px 0;'><span style='color: #32CD32; font-weight: bold; font-size: 20px;'>●</span> <strong>Green</strong> - Cells infected by DIPs only</li>",
        "<li style='margin-bottom: 4px; padding: 3px 0;'><span style='color: #FF4500; font-weight: bold; font-size: 20px;'>●</span> <strong>Red</strong> - Cells infected by virions only</li>",
        "<li style='margin-bottom: 4px; padding: 3px 0;'><span style='color: #4169E1; font-weight: bold; font-size: 20px;'>●</span> <strong>Blue</strong> - Cells in antiviral state</li>",
        "<li style='margin-bottom: 4px; padding: 3px 0;'><span style='color: #800080; font-weight: bold; font-size: 20px;'>●</span> <strong>Purple</strong> - Regrowth cells</li>",
        "</ul>",
        "</div>",
        "</div>"
      )
    } else {
      # Fallback to video if GIF creation fails or no PNG exists
      if (file.exists(mp4_file)) {
        # Copy video to www folder for web access
        www_dir <- "www"
        if (!dir.exists(www_dir)) dir.create(www_dir)
        
        # Check if file is actually AVI format
        file_info <- system2("file", args = mp4_file, stdout = TRUE)
        if (grepl("AVI", file_info)) {
          # If it's AVI, rename it properly
          mp4_name <- paste0("video_", round(as.numeric(Sys.time())), ".avi")
          mp4_path <- file.path(www_dir, mp4_name)
          file.copy(mp4_file, mp4_path, overwrite = TRUE)
        } else {
          # If it's actually MP4, keep original name
          mp4_name <- paste0("video_", round(as.numeric(Sys.time())), ".mp4")
          mp4_path <- file.path(www_dir, mp4_name)
          file.copy(mp4_file, mp4_path, overwrite = TRUE)
        }
        
        # Determine video type and appropriate HTML
        file_ext <- tools::file_ext(mp4_name)
        video_type <- if (file_ext == "avi") "video/avi" else "video/mp4"
        
        html_content <- paste0(html_content,
          "<p><strong>Simulation Video:</strong></p>",
          "<div style='text-align: center; margin: 20px 0; background: #f8f9fa; padding: 20px; border-radius: 8px; border: 1px solid #ddd;'>",
          "<video controls preload='metadata' style='width: 80%; min-width: 400px; max-width: 800px; height: auto; border: 2px solid #007bff; border-radius: 8px; box-shadow: 0 4px 8px rgba(0,0,0,0.1);'>",
          "<source src='www/", mp4_name, "' type='", video_type, "'>",
          "<source src='www/", mp4_name, "' type='video/x-msvideo'>",
          "Your browser does not support the video tag. <a href='www/", mp4_name, "' download style='color: #007bff; text-decoration: underline;'>Download video</a>",
          "</video>",
          "<br/><br/>",
          "<p style='color: #666; font-size: 14px; margin-top: 10px;'>Video format: 400x400 pixels, Motion JPEG (", toupper(file_ext), ")</p>",
          "<p style='color: #666; font-size: 14px;'>If video doesn't play, try downloading it or use VLC media player.</p>",
          "</div>"
        )
      }
    }
    
    # Add image if it exists
    if (file.exists(png_file)) {
      # Copy image to www folder for web access
      www_dir <- "www"
      if (!dir.exists(www_dir)) dir.create(www_dir)
      img_name <- paste0("result_", round(as.numeric(Sys.time())), ".png")
      img_path <- file.path(www_dir, img_name)
      file.copy(png_file, img_path, overwrite = TRUE)
      
      html_content <- paste0(html_content,
        "<p><strong>Simulation Visualization at Different Time Points:</strong></p>",
        "<div style='text-align: center; margin: 10px 0;'>",
        "<img src='www/", img_name, "' alt='Simulation Results' style='max-width: 100%; height: auto; border: 1px solid #ddd; border-radius: 4px;'>",
        "</div>"
      )
    }
    
    # Add download buttons after image
    html_content <- paste0(html_content,
      "<p><strong>Download Files:</strong></p>",
      "<div style='margin: 15px 0; display: flex; gap: 10px; flex-wrap: wrap;'>"
    )
    
    # Add video download if available
    if (file.exists(mp4_file)) {
      # Use the same video file name that was created earlier for display
      existing_videos <- list.files("www", pattern = "^video_.*\\.(mp4|avi)$", full.names = FALSE)
      if (length(existing_videos) > 0) {
        video_name <- existing_videos[1]  # Use the first/most recent video file
      } else {
        # Fallback if no existing video found
        video_name <- paste0("video_", round(as.numeric(Sys.time())), ".mp4")
        video_path <- file.path("www", video_name)
        file.copy(mp4_file, video_path, overwrite = TRUE)
      }
      
      # Get file extension for proper download name
      file_ext <- tools::file_ext(video_name)
      download_name <- paste0("simulation_video.", file_ext)
      
      html_content <- paste0(html_content,
        "<a href='www/", video_name, "' download='", download_name, "' style='flex: 1; display: block; padding: 20px 25px; background: #007bff; color: white; text-decoration: none; border-radius: 8px; text-align: center; font-size: 24px; font-weight: bold;'>📹 Download Video (", toupper(file_ext), ")</a>"
      )
    }
    
    if (file.exists(csv_file)) {
      # Copy CSV to www folder for download
      www_dir <- "www"
      csv_name <- paste0("data_", round(as.numeric(Sys.time())), ".csv")
      csv_path <- file.path(www_dir, csv_name)
      file.copy(csv_file, csv_path, overwrite = TRUE)
      
      html_content <- paste0(html_content,
        "<a href='www/", csv_name, "' download='simulation_data.csv' style='flex: 1; display: block; padding: 20px 25px; background: #28a745; color: white; text-decoration: none; border-radius: 8px; text-align: center; font-size: 24px; font-weight: bold;'>📊 Download Data (CSV)</a>"
      )
    }
    
    html_content <- paste0(html_content, "</div>")
    
    # Add CSV data preview if it exists
    if (file.exists(csv_file)) {
      tryCatch({
        # Read CSV with more specific settings
        csv_data <- read.csv(csv_file, nrows = 10, stringsAsFactors = FALSE, check.names = FALSE)
        
        if (nrow(csv_data) > 0 && ncol(csv_data) > 0) {
          html_content <- paste0(html_content,
            "<p><strong>CSV Data Preview (First ", min(nrow(csv_data), 5), " rows of ", ncol(csv_data), " columns):</strong></p>",
            "<div style='overflow-x: auto; margin: 10px 0; border: 1px solid #ddd; border-radius: 4px;'>",
            "<table style='border-collapse: collapse; width: 100%; font-size: 11px; margin: 0;'>",
            "<thead>",
            "<tr style='background: #f8f9fa;'>"
          )
          
          # Add headers with better formatting
          for (col in names(csv_data)) {
            clean_col <- gsub("'", "\\'", col)
            html_content <- paste0(html_content, "<th style='border: 1px solid #ddd; padding: 6px; text-align: left; font-weight: bold; word-wrap: break-word; max-width: 120px;'>", clean_col, "</th>")
          }
          html_content <- paste0(html_content, "</tr></thead><tbody>")
          
          # Add data rows with better formatting
          for (i in 1:min(nrow(csv_data), 5)) {
            html_content <- paste0(html_content, "<tr", ifelse(i %% 2 == 0, " style='background: #f9f9f9;'", ""), ">")
            for (col in names(csv_data)) {
              value <- csv_data[i, col]
              # Format the value
              if (is.numeric(value)) {
                if (abs(value) < 0.001 && value != 0) {
                  formatted_value <- sprintf("%.2e", value)
                } else {
                  formatted_value <- round(value, 3)
                }
              } else {
                formatted_value <- as.character(value)
                formatted_value <- gsub("'", "\\'", formatted_value)
              }
              html_content <- paste0(html_content, "<td style='border: 1px solid #ddd; padding: 6px; word-wrap: break-word; max-width: 120px;'>", formatted_value, "</td>")
            }
            html_content <- paste0(html_content, "</tr>")
          }
          html_content <- paste0(html_content, "</tbody></table></div>")
        } else {
          html_content <- paste0(html_content, "<p><strong>CSV:</strong> File exists but appears to be empty</p>")
        }
      }, error = function(e) {
        html_content <- paste0(html_content, "<p><strong>CSV Error:</strong> ", e$message, "</p>")
      })
    }
    
    # Close the results div and add parameters at the bottom
    html_content <- paste0(html_content, "</div>")
    
    # Add parameters information at the bottom
    html_content <- paste0(html_content,
      "<div style='text-align: left; background: #f8f9fa; padding: 15px; border-radius: 8px; margin: 10px 0; border-left:4px solid #28a745; color: #333333;'>",
      "<h5 style='color:#28a745;'>Parameters Used:</h5>",
      "<ul style='line-height:1.7;'>",
      "<li>Tau: ", input$tau, "</li>",
      "<li>Burst Size V: ", input$burstSizeV, "</li>",
      "<li>Burst Size D: ", if(input$dipOption) input$burstSizeD else "(Disabled)", "</li>",
      "<li>Mean Lysis Time: ", input$meanLysisTime, "</li>",
      "<li>Rho: ", input$rho, "</li>",
      "<li>DIP Half Life: ", if(input$dipOption) input$dip_half_life else "(Disabled)", "</li>",
      "<li>DIP Synthesis Advantage: ", if(input$dipOption) input$dipSynthesisAdvantage else "(Disabled)", "</li>",
      "<li>Option: ", option_value, "</li>",
      "<li>Video Type: ", input$videotype, "</li>",
      "<li>Particle Spread: ", input$particleSpreadOption, "</li>",
      "<li>IFN Spread: ", input$ifnSpreadOption, "</li>",
      "<li>DIP Option: ", ifelse(input$dipOption, "Enabled", "Disabled"), "</li>",
      "</ul>",
      "</div>"
    )
    
    return(html_content)
  } else {
    return(paste0("
      <h4 style='color:#ff8c00; margin-left: 40px;'>Simulation Complete</h4>
      <div style='background: #fff3e0; padding: 15px; border-radius: 8px; margin: 10px 0; border-left:4px solid #ff8c00; color: #333333;'>
        <p>Script executed but no output folder found.</p>
      </div>
    "))
  }
}

# Define server logic
server <- function(input, output, session) {
  # Create www directory if it doesn't exist and add resource path
  if (!dir.exists("www")) {
    dir.create("www")
  }
  addResourcePath("www", "www")
  
  # Set default option value to 1 (since it's no longer in UI)
  option_value <- 1
  
  # Observe DIP Option changes to enable/disable related inputs
  observe({
    if (input$dipOption) {
      shinyjs::enable("burstSizeD")
      shinyjs::enable("dip_half_life")
      shinyjs::enable("dipSynthesisAdvantage")
    } else {
      shinyjs::disable("burstSizeD")
      shinyjs::disable("dip_half_life")
      shinyjs::disable("dipSynthesisAdvantage")
    }
  })
  
  observeEvent(input$run_simulation, {
    # Show progress display area and bottom-right status
    runjs("document.getElementById('progress_display').style.display = 'block';")
    runjs("document.getElementById('progress_status').style.display = 'block';")
    # Clear previous results and show starting message
    runjs("document.getElementById('simulation_results').innerHTML = '<p style=\"margin-left: 40px;\">Simulation in progress...</p>';")
    
    # Build command line arguments for Go script
    go_args <- c(
      paste0("-tau=", input$tau),
      paste0("-burstSizeV=", input$burstSizeV),
      paste0("-meanLysisTime=", input$meanLysisTime),
      paste0("-rho=", input$rho),
      paste0("-virion_half_life=", input$virion_half_life),
      paste0("-ifn_half_life=", input$ifn_half_life),
      paste0("-particleSpreadOption=", input$particleSpreadOption),
      paste0("-ifnSpreadOption=", input$ifnSpreadOption),
      paste0("-dipOption=", tolower(input$dipOption)),
      paste0("-videotype=", input$videotype),
      paste0("-option=", option_value)
    )
    
    # Add DIP-specific parameters only if dipOption is enabled
    if (input$dipOption) {
      go_args <- c(go_args,
        paste0("-burstSizeD=", input$burstSizeD),
        paste0("-dip_half_life=", input$dip_half_life),
        paste0("-dipSynthesisAdvantage=", input$dipSynthesisAdvantage),
        "-d_pfu_initial=1.0"
      )
    }
    
    # Handle kJumpR parameter
    if (input$kJumpR == "NaN") {
      go_args <- c(go_args, "-kJumpR=0.0")
    } else {
      go_args <- c(go_args, paste0("-kJumpR=", input$kJumpR))
    }
    
    # Add ifnBothFold parameter
    go_args <- c(go_args, paste0("-ifnBothFold=", input$ifnBothFold))
    
    # Simulate progress up to 99% (much slower progression)
    withProgress(message = 'Simulation in Progress', value = 0, {
      for (i in 1:99) {
        incProgress(1/99, detail = paste0('Completed ', i, '%'))
        runjs(paste0("document.getElementById('progress_status').innerHTML = 'Progress: ", i, "%';"))
        runjs(paste0("document.getElementById('progress_text').innerHTML = '<h5 style=\"margin-left: 40px;\">Simulation Running...</h5><p style=\"margin-left: 40px;\">Progress: ", i, "%</p><p style=\"margin-left: 40px;\">Processing parameters...</p>';"))
        Sys.sleep(0.4)  # Even slower: 0.4 seconds per percent
      }
      
      # Stop at 99% and start Go script execution
      runjs("document.getElementById('progress_status').innerHTML = 'Progress: 99% - Running Script...';")
      runjs("document.getElementById('progress_text').innerHTML = '<h5 style=\"margin-left: 40px;\">Simulation at 99%</h5><p style=\"margin-left: 40px;\">waiting for R Shiny website signal...</p><p style=\"margin-left: 40px;\">please wait for a few seconds...</p>';")
      
      # Execute Go script
      tryCatch({
        result <- system2("go", args = c("run", "mdbk_small_vero_0716.go", go_args), 
                         stdout = TRUE, stderr = TRUE, wait = TRUE)
        
        if (attr(result, "status") == 0 || is.null(attr(result, "status"))) {
          # Go script completed successfully - ResOut condition met!
          runjs("document.getElementById('progress_status').innerHTML = 'Progress: 100% - Complete!';")
          runjs("document.getElementById('progress_text').innerHTML = '<h5 style=\"color: #28a745; margin-left: 40px;\">Simulation Complete!</h5><p style=\"margin-left: 40px;\">Executed successfully.</p>';")
          
          # Generate results display
          results_html <- generate_results_display(input, option_value, result)
          # Escape quotes and handle special characters for JavaScript
          results_html_escaped <- gsub("'", "\\\\'", results_html)
          results_html_escaped <- gsub("\n", " ", results_html_escaped)
          results_html_escaped <- gsub("\r", " ", results_html_escaped)
          runjs(paste0("document.getElementById('simulation_results').innerHTML = '", results_html_escaped, "';"))
          
        } else {
          # Go script failed
          runjs("document.getElementById('progress_status').innerHTML = 'Error: Go script failed';")
          runjs("document.getElementById('progress_text').innerHTML = '<h5 style=\"color: #dc3545;\">Error!</h5><p>Script execution failed.</p>';")
          
          error_html <- paste0("
            <h4 style='color:#dc3545;'>Simulation Failed</h4>
            <div style='background: #f8d7da; padding: 15px; border-radius: 8px; margin: 10px 0; border-left:4px solid #dc3545; color: #721c24;'>
              <h5>Error Details:</h5>
              <pre>", paste(result, collapse = "\\n"), "</pre>
            </div>
          ")
          runjs(paste0("document.getElementById('simulation_results').innerHTML = '", error_html, "';"))
        }
      }, error = function(e) {
        # R error in execution
        runjs("document.getElementById('progress_status').innerHTML = 'Error: Execution failed';")
        runjs("document.getElementById('progress_text').innerHTML = '<h5 style=\"color: #dc3545;\">Error!</h5><p>Failed to execute Go script.</p>';")
        
        error_html <- paste0("
          <h4 style='color:#dc3545;'>Execution Error</h4>
          <div style='background: #f8d7da; padding: 15px; border-radius: 8px; margin: 10px 0; border-left:4px solid #dc3545; color: #721c24;'>
            <h5>Error Details:</h5>
            <p>", e$message, "</p>
          </div>
        ")
        runjs(paste0("document.getElementById('simulation_results').innerHTML = '", error_html, "';"))
      })
    })
    
    # Hide progress displays after 5 seconds
    runjs("setTimeout(function() { 
      document.getElementById('progress_status').style.display = 'none'; 
      document.getElementById('progress_display').style.display = 'none'; 
    }, 5000);")
  })
}

# Run the application
shinyApp(ui = ui, server = server)
