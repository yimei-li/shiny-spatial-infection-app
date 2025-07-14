# GIF Generator for Simulation Results
# This script converts the selected_frames_combined.png into an animated GIF

# Set CRAN mirror
options(repos = c(CRAN = "https://cloud.r-project.org/"))

# Install and load required packages
if (!requireNamespace("magick", quietly = TRUE)) {
  install.packages("magick")
}
library(magick)

# Function to create GIF from combined frames image
create_gif_from_combined_image <- function(input_image_path, output_gif_path, num_frames = 20) {
  tryCatch({
    # Check if input file exists
    if (!file.exists(input_image_path)) {
      stop("Input image file not found: ", input_image_path)
    }
    
    # Read the combined image
    combined_image <- image_read(input_image_path)
    
    # Get image dimensions
    info <- image_info(combined_image)
    total_width <- info$width
    frame_height <- info$height
    
    # Calculate frame width (should be total_width / num_frames)
    frame_width <- total_width / num_frames
    
    # Adjust dimensions to remove black borders
    # Remove 100 pixels from right and 53 pixels from bottom
    cropped_width <- frame_width - 100
    cropped_height <- frame_height - 53
    
    cat("Image info:\n")
    cat("  Total width:", total_width, "pixels\n")
    cat("  Frame height:", frame_height, "pixels\n")
    cat("  Frame width:", frame_width, "pixels\n")
    cat("  Cropped width:", cropped_width, "pixels\n")
    cat("  Cropped height:", cropped_height, "pixels\n")
    cat("  Number of frames:", num_frames, "\n")
    
    # Create a list to store individual frames
    frames <- list()
    
    # Extract each frame by cropping
    for (i in 1:num_frames) {
      # Calculate the x offset for this frame
      x_offset <- (i - 1) * frame_width
      
      # Define the crop geometry: widthxheight+x_offset+y_offset
      # Remove black borders by adjusting width and height
      crop_geometry <- paste0(cropped_width, "x", cropped_height, "+", x_offset, "+0")
      
      # Crop the frame
      frame <- image_crop(combined_image, crop_geometry)
      
      # Add frame to list
      frames[[i]] <- frame
      
      cat("Extracted frame", i, "at offset", x_offset, "(", cropped_width, "x", cropped_height, ")\n")
    }
    
    # Combine all frames into an animated image
    animated_image <- image_join(frames)
    
    # Set animation properties
    # animate() sets the delay between frames (in 1/100th of a second)
    # A delay of 50 means 0.5 seconds between frames
    animated_gif <- image_animate(animated_image, fps = 2, dispose = "previous")
    
    # Write the GIF
    image_write(animated_gif, output_gif_path)
    
    cat("GIF created successfully:", output_gif_path, "\n")
    cat("Animation: ", num_frames, " frames at 2 FPS\n")
    
    return(TRUE)
    
  }, error = function(e) {
    cat("Error creating GIF:", e$message, "\n")
    return(FALSE)
  })
}

# Function to create GIF from simulation output folder
create_gif_from_simulation_folder <- function(folder_path, output_folder = "www") {
  # Look for the combined image file
  combined_image_path <- file.path(folder_path, "selected_frames_combined.png")
  
  if (!file.exists(combined_image_path)) {
    cat("No combined image found in:", folder_path, "\n")
    return(FALSE)
  }
  
  # Create output directory if it doesn't exist
  if (!dir.exists(output_folder)) {
    dir.create(output_folder, recursive = TRUE)
  }
  
  # Generate output GIF path
  timestamp <- round(as.numeric(Sys.time()))
  output_gif_path <- file.path(output_folder, paste0("animation_", timestamp, ".gif"))
  
  # Create the GIF
  success <- create_gif_from_combined_image(combined_image_path, output_gif_path)
  
  if (success) {
    return(output_gif_path)
  } else {
    return(NULL)
  }
}

# Test function (can be called manually)
test_gif_creation <- function() {
  # Find the most recent simulation folder
  output_folders <- list.dirs(".", recursive = FALSE, full.names = TRUE)
  output_folders <- output_folders[sapply(output_folders, function(folder) {
    file.exists(file.path(folder, "selected_frames_combined.png"))
  })]
  
  if (length(output_folders) > 0) {
    # Get the most recent folder
    latest_folder <- output_folders[which.max(file.info(output_folders)$mtime)]
    cat("Testing GIF creation with folder:", latest_folder, "\n")
    
    result <- create_gif_from_simulation_folder(latest_folder)
    if (!is.null(result)) {
      cat("Test successful! GIF created at:", result, "\n")
    } else {
      cat("Test failed!\n")
    }
  } else {
    cat("No simulation folders found with combined images.\n")
  }
} 