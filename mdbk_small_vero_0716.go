/*
 * Author: Yimei Li
 * Affiliation: Princeton University, Grenfell Lab / teVelthuis Lab / Levin Lab
 * Year: 2024
 * Copyright: © 2024 Yimei Li. All rights reserved.
 * License: Proprietary. All rights reserved.
 *
 * Usage: Used to generate IFN dynamics plots and summary statistics for simulations in my PhD thesis.
 *
 * NOTE: Burst size is fixed. The burst size for virions (burst size V) is a constant value in this simulation.
 */

//: A simpler script with a higher-resolution image for easier inspection. I changed the GRID_SIZE to a smaller value, like 10, and set HOWAT_V_PFU_INITIA to 1, a small initial number of virions. I also increased CELL_SIZE to 10 for a clearer image. Additionally, I changed the "option" from 3 to 2 so you can modify the initial virion location. In option 2, I set the initial location at [4][5] ( "g.localVirions[4][5]++" ) which you can change. RHO is 1, and the virus spreads to neighboring cells weighted by distance.

package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os" // Used for file operations
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/icza/mjpeg"
	"github.com/wcharczuk/go-chart/v2" // Used for plotting the graph
	"github.com/wcharczuk/go-chart/v2/drawing"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Constant definitions
const (
	TIME_STEPS = 502 // Number of time steps
	GRID_SIZE  = 50  // Size of the grid

	FRAME_RATE   = 1          // Frame rate for the video
	OUTPUT_VIDEO = "0421.mp4" // Output video file name
	CELL_SIZE    = 4          // The size of each hexagonal cell
	TIMESTEP     = 1          // Time step size

)

// Define flag variables (note they are all pointer types)
var (

	// Option parameters
	// Particle spread option: can be "celltocell", "jumprandomly", "partition" or "jumpradius"
	flag_particleSpreadOption = flag.String("particleSpreadOption", "jumprandomly", "Particle spread option: celltocell, jumprandomly, jumpradius, or partition")
	// If jumprandomly is selected, this parameter represents the random jump ratio (0~1)

	// IFN spread option: can be "global", "local" or "noIFN"
	flag_ifnSpreadOption = flag.String("ifnSpreadOption", "local", "IFN spread option: global, local, or noIFN")
	// DIP option: if true then enable DIP, if false then disable DIP
	flag_dipOption = flag.Bool("dipOption", true, "DIP option: if true then enable DIP, if false then disable DIP")

	flag_burstSizeV       = flag.Int("burstSizeV", 50, "Number of virions released when a cell lyses")
	flag_burstSizeD       = flag.Int("burstSizeD", 100, "Number of DIPs released when a cell lyses")
	flag_meanLysisTime    = flag.Float64("meanLysisTime", 12.0, "Mean lysis time")
	flag_kJumpR           = flag.Float64("kJumpR", 0.5, "Parameter for cell-to-cell jump randomness")
	flag_tau              = flag.Int("tau", 12, "TAU value (e.g., lysis time)")
	flag_ifnBothFold      = flag.Float64("ifnBothFold", 1.0, "Fold effect for IFN stimulation")
	flag_rho              = flag.Float64("rho", 0.026, "Infection rate constant")
	flag_virion_half_life = flag.Float64("virion_half_life", 3.2, "Virion clearance rate (e.g., 3.2 d^-1)")
	flag_dip_half_life    = flag.Float64("dip_half_life", 3.2, "DIP clearance rate (e.g., 3.2 d^-1)")
	flag_ifn_half_life    = flag.Float64("ifn_half_life", 4.0, "IFN clearance rate (e.g., 3.0 d^-1)")
	flag_option           = flag.Int("option", 2, "Option for infection initialization (e.g., 1, 2, 3)")

	flag_v_pfu_initial = flag.Float64("v_pfu_initial", 1.0, "Initial PFU count for virions")
	flag_d_pfu_initial = flag.Float64("d_pfu_initial", 0.0, "Initial PFU count for DIPs")
	flag_videotype     = flag.String("videotype", "states", "Video type: states, IFNconcentration, IFNonlyLargerThanZero, antiviralState, particles")
)

// Particle spread related
var (
	particleSpreadOption  string  // "celltocell", "jumprandomly", "jumpradius"
	jumpRadiusV           int     // e.g., when "jumpradius" is selected, set to 5
	jumpRadiusD           int     // same as above
	jumpRandomly          bool    // whether to use random jump (true when "jumprandomly" is selected)
	k_JumpR               float64 // random jump ratio
	par_celltocell_random bool
)

// IFN spread related
var (
	ifnSpreadOption string // "global",
	//  "local", "noIFN"
	IFN_wave_radius int  // if ifnSpreadOption=="local", e.g., set to 10; "global" or "noIFN" set to 0
	ifnWave         bool // whether to enable IFN wave
)

// DIP related
var (
	dipOption bool // true to enable DIP, false to disable DIP

	// When DIP is enabled, default DIP-related ratios remain default; when disabled, set to 0
	D_only_IFN_stimulate_ratio float64 = 5.0 * ifnBothFold
	BOTH_IFN_stimulate_ratio   float64 = 10.0 * ifnBothFold
)

// Global variables
var (
	// particleSpreadOption  = "jumpradius" // options: "celltocell", "jumprandomly", "jumpradius"
	// PartionParticleSpreadOption = false // options: "true" or "false"

	//ifnSpreadOption = "global" // options: "global", "local" or "noIFN"
	//dipOption =

	BURST_SIZE_V  int    // CHANGE 50 Number of virions released when a cell lyses
	BURST_SIZE_D  int    // CHANGE 100 // Number of DIPs released when a cell lyses
	VStimulateIFN = true // CHANGE if false then usually only DIP stimulate IFN in this situlation, not virion
	//jumpRandomly          = true // CHANGE
	//jumpRadiusV           = 0    // CHANGE Virion jump radius
	//jumpRadiusD           = 0    // CHANGE DIP jump radius
	//IFN_wave_radius       = 10   // CHANGE 10
	// this is true only when jumpRandomly is true

	TAU         int // 95
	ifnBothFold = 1.0
	//D_only_IFN_stimulate_ratio = 5.0 * ifnBothFold  // D/V *R *D_only_IFN_stimulate_ratio
	//BOTH_IFN_stimulate_ratio = 10.0 * ifnBothFold // D/V *R *D_only_IFN_stimulate_ratio

	// option    = 2        // Option for infection initialization
	//videotype = "states" // "states" // color in "states" or "IFNconcentration" or "IFNonlyLargerThanZero" or "antiviralState" or "particles"
	RHO    float64 //0.026    //0.02468  // 0.09 Infection rate constant
	option int
	// radius 10 of grid has 331 cells
	R int
	// radius 10 of grid has 331 cells, originally infected cell increases R IFN,
	ALPHA = 1.0 // Parameter for infection probability (set to 1.5)

	REGROWTH_MEAN       = 24.0  // Mean time for regrowth
	REGROWTH_STD        = 6.0   // Standard deviation for regrowth time
	MEAN_LYSIS_TIME     float64 // Mean lysis time
	STANDARD_LYSIS_TIME float64 // Standard deviation for lysis time
	maxGlobalIFN        = -1.0  // used to track maximum IFN value
	globalIFN           = -1.0  // global IFN concentration
	globalIFNperCell    = 0.0
	IFN_DELAY           = 5
	STD_IFN_DELAY       = 1

	// allowVirionJump = jumpRadiusV > 0 || jumpRandomly // Allow virions to jump to other cells
	// allowDIPJump = jumpRadiusD > 0 || jumpRandomly // Allow DIPs to jump to other cells
	allowVirionJump bool
	allowDIPJump    bool
	//ifnWave = IFN_wave_radius > 0

	yMax          float64
	xMax          = float64(TIME_STEPS)
	ticksInterval float64 // Interval for X-axis ticks

	adjusted_DIP_IFN_stimulate   float64
	perParticleInfectionChance_V float64
	totalDeadFromBoth            int
	totalDeadFromV               int
	virionDiffusionRate          int
	dipDiffusionRate             int

	virion_half_life float64 //= 0.0 // 3.2 // ~4 d^-1 => half-life ~4.2 hours
	dip_half_life    float64 //= 0.0 // 3.2 // ~4 d^-1 => half-life ~4.2 hours
	ifn_half_life    float64 //= 0.0 // 3.0 // ~3 d^-1 => half-life ~5.5 hours
	videotype        string
	dipAdvantage     float64 // DIP advantage = burstSizeD / burstSizeV
)

// Cell state definitions
const (
	SUSCEPTIBLE     = 0 // Susceptible state
	INFECTED_VIRION = 1 // Infected by virion
	INFECTED_DIP    = 5 // Infected by DIP
	INFECTED_BOTH   = 6 // Infected by both virion and DIP
	DEAD            = 2 // Dead state
	ANTIVIRAL       = 3 // Antiviral state
	REGROWTH        = 4 // Regrowth state
)

// Grid structure for storing the simulation state
type Grid struct {
	state                  [GRID_SIZE][GRID_SIZE]int        // State of the cells in the grid
	localVirions           [GRID_SIZE][GRID_SIZE]int        // Number of virions in each cell
	localDips              [GRID_SIZE][GRID_SIZE]int        // Number of DIPs in each cell
	IFNConcentration       [GRID_SIZE][GRID_SIZE]float64    // IFN concentration in each cell
	timeSinceInfectVorBoth [GRID_SIZE][GRID_SIZE]int        // Time since infection for each cell
	timeSinceInfectDIP     [GRID_SIZE][GRID_SIZE]int        // Time since infection for each cell
	timeSinceDead          [GRID_SIZE][GRID_SIZE]int        // Time since death for each cell
	timeSinceRegrowth      [GRID_SIZE][GRID_SIZE]int        // Time since regrowth for each cell
	timeSinceSusceptible   [GRID_SIZE][GRID_SIZE]int        // Time since cell became susceptible
	neighbors1             [GRID_SIZE][GRID_SIZE][6][2]int  // Neighbors at distance 1
	neighbors2             [GRID_SIZE][GRID_SIZE][6][2]int  // Neighbors at distance 2
	neighbors3             [GRID_SIZE][GRID_SIZE][6][2]int  // Neighbors at distance 3
	neighborsRingVirion    [GRID_SIZE][GRID_SIZE][60][2]int // Neighbors at distance 10 ring
	neighborsRingDIP       [GRID_SIZE][GRID_SIZE][60][2]int // Neighbors at distance 10 ring
	neighborsIFNArea       [GRID_SIZE][GRID_SIZE][][2]int   // Neighbors within IFN wave radius
	stateChanged           [GRID_SIZE][GRID_SIZE]bool       // Flag to indicate if the state of a cell has changed
	antiviralDuration      [GRID_SIZE][GRID_SIZE]int        // Duration of antiviral state
	previousStates         [GRID_SIZE][GRID_SIZE]int        // Previous state of the cell
	antiviralFlag          [GRID_SIZE][GRID_SIZE]bool       // Flag to indicate if the cell is in the antiviral state
	timeSinceAntiviral     [GRID_SIZE][GRID_SIZE]int        // Time since the cell entered the antiviral state
	antiviralCellCount     int                              // Number of cells in the antiviral state
	totalAntiviralTime     int
	intraWT                [GRID_SIZE][GRID_SIZE]int // IntraWT
	intraDVG               [GRID_SIZE][GRID_SIZE]int // IntraDVG
	allowJumpRandomly      [][]bool
	totalRandomJumpVirions int                       // record total number of randomly jumping Virions
	totalRandomJumpDIPs    int                       // record total number of randomly jumping DIPs
	lysisThreshold         [GRID_SIZE][GRID_SIZE]int // fixed lysis time for each cell

}

// Initialize the infection state
func (g *Grid) initializeInfection(option int) {
	// Use current time as seed for reproducibility
	rand.Seed(time.Now().UnixNano())

	vInit := int(math.Round(*flag_v_pfu_initial))
	dInit := int(math.Round(*flag_d_pfu_initial))

	switch option {
	case 1:
		if vInit > 0 {
			g.localVirions[25][25] = vInit
		} else {
			fmt.Printf("v_pfu_initial < 0: %.2f\n", *flag_v_pfu_initial)
		}
		if dInit > 0 {
			g.localDips[25][25] = dInit
		} else {
			fmt.Printf("d_pfu_initial < 0: %.2f\n", *flag_d_pfu_initial)
		}
	case 2:
		if vInit > 0 && dInit > 0 {
			g.state[25][25] = INFECTED_BOTH
		} else if vInit > 0 {
			g.state[25][25] = INFECTED_VIRION
		} else if dInit > 0 {
			g.state[25][25] = INFECTED_DIP
		}
		g.localVirions[25][25] = vInit
		g.localDips[25][25] = dInit

	case 3:
		for k := 0; k < vInit; k++ {
			i := rand.Intn(GRID_SIZE)
			j := rand.Intn(GRID_SIZE)
			g.localVirions[i][j]++
		}
		for k := 0; k < dInit; k++ {
			i := rand.Intn(GRID_SIZE)
			j := rand.Intn(GRID_SIZE)
			g.localDips[i][j]++
		}
	}
}

// Function to generate ticks dynamically
func generateTicks(xMax float64, interval float64) []chart.Tick {
	var ticks []chart.Tick
	for value := 0.0; value <= xMax; value += interval {
		label := fmt.Sprintf("%.0f", value) // Format the label as an integer
		ticks = append(ticks, chart.Tick{
			Value: value,
			Label: label,
		})
	}
	return ticks
}

// Initialize the grid, setting all cells to SUSCEPTIBLE
func (g *Grid) initialize() {
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			g.state[i][j] = SUSCEPTIBLE
			g.stateChanged[i][j] = false // Initialize as unchanged
			g.timeSinceInfectVorBoth[i][j] = -1
			g.timeSinceDead[i][j] = -1
			g.timeSinceRegrowth[i][j] = -1
			g.IFNConcentration[i][j] = 0
			g.antiviralDuration[i][j] = -1
			g.timeSinceSusceptible[i][j] = 0
			g.previousStates[i][j] = -1
			g.antiviralFlag[i][j] = false
			g.timeSinceAntiviral[i][j] = -1
			g.intraWT[i][j] = 0
			g.intraDVG[i][j] = 0
			g.lysisThreshold[i][j] = -1

		}
	}

	fmt.Println("Grid initialized")

}

// Ensure the entire canvas is initialized with uniform background color
func fillBackground(img *image.RGBA, bgColor color.Color) {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, bgColor)
		}
	}
}

func (g *Grid) calculateDiffusionRates() (float64, float64) {
	totalVirions := 0
	totalVirionDiffusion := 0
	totalDIPs := 0
	totalDIPDiffusion := 0

	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			virionMoves := g.localVirions[i][j]
			dipMoves := g.localDips[i][j]

			// Sum particles moved out for diffusion
			totalVirionDiffusion += virionMoves
			totalDIPDiffusion += dipMoves

			// Add to total particles in the grid
			totalVirions += g.localVirions[i][j]
			totalDIPs += g.localDips[i][j]
		}
	}

	virionDiffusionRate := float64(totalVirionDiffusion) / float64(totalVirions)
	dipDiffusionRate := float64(totalDIPDiffusion) / float64(totalDIPs)
	return virionDiffusionRate, dipDiffusionRate
}

// Function to get the nth figure number in the folder
func getNextFigureNumber(outputFolder string) int {
	files, err := os.ReadDir(outputFolder)
	if err != nil {
		log.Fatalf("Failed to read output folder: %v", err)
	}
	count := 0
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".png") {
			count++
		}
	}
	return count + 1 // Return the next number
}

// Logic to determine IFN spreading type
func getIFNType() string {
	if IFN_wave_radius == 0 && globalIFN > 0 {
		return "Global"
	} else if IFN_wave_radius == 10 {
		return "IFNJ10"
	} else {
		return "NoIFN"
	}
}

func generateFolderName(
	no int,
	jumpRandomly bool,
	jumpRadiusD int,
	jumpRadiusV int,
	burstSizeD int,
	burstSizeV int,
	ifnWaveRadius int,
	TAU int,
	timeSteps int,
) string {
	// Determine Dinit naming part (keep at most 2 decimal places)
	dInit := fmt.Sprintf("Dinit%s", strconv.FormatFloat(*flag_d_pfu_initial, 'f', -1, 64))

	// Determine D naming part
	dName := ""
	if jumpRandomly {
		dName = fmt.Sprintf("DIPBst%d_JRand", burstSizeD)
	} else if jumpRadiusD > 0 {
		dName = fmt.Sprintf("DIPBst%d_J%d", burstSizeD, jumpRadiusD)
	} else if jumpRadiusD == 0 {
		dName = fmt.Sprintf("DIPBst%d_noJ", burstSizeD)
	} else {
		if burstSizeD == 0 && *flag_d_pfu_initial == 0 && D_only_IFN_stimulate_ratio == 0 && jumpRadiusD == 0 {
			dName = "NoDIP"
		} else {
			dName = fmt.Sprintf("DIPBst%d", burstSizeD)
		}
	}

	// Determine Vinit naming part (keep at most 2 decimal places)
	vInit := ""
	if *flag_v_pfu_initial > 0 {
		vInit = fmt.Sprintf("Vinit%s", strconv.FormatFloat(*flag_v_pfu_initial, 'f', -1, 64))
	} else if jumpRandomly {
		vInit = "JRand"
	} else if jumpRadiusV > 0 {
		vInit = fmt.Sprintf("J%d", jumpRadiusV)
	} else {
		vInit = "noJ"
	}

	vName := fmt.Sprintf("VBst%d", burstSizeV)

	// Determine IFN naming part
	ifnName := ""
	if TAU == 0 {
		ifnName = "NoIFN"
	} else if ifnWaveRadius == 0 {
		ifnName = "Global"
	} else {
		ifnName = fmt.Sprintf("IFN%d", ifnWaveRadius)
	}

	cellType := ""
	if TAU > 0 {
		cellType = "mdbk"
	} else {
		cellType = "vero"
	}

	folderName := fmt.Sprintf("%d_%s_%s_%s_%s_%s_%s_times%d_tau%d_ifnBothFold%.2f_grid%d_VStimulateIFN%t",
		no, dInit, dName, vInit, vName, ifnName, cellType, timeSteps, TAU, ifnBothFold, GRID_SIZE, VStimulateIFN)

	return folderName
}

// Combine images into one row
func combineImagesHorizontally(images []*image.RGBA) *image.RGBA {
	if len(images) == 0 {
		return nil
	}

	// Calculate the width and height of the combined image
	totalWidth := 0
	maxHeight := 0
	for _, img := range images {
		totalWidth += img.Bounds().Dx() // accumulate width
		if img.Bounds().Dy() > maxHeight {
			maxHeight = img.Bounds().Dy() // calculate maximum height
		}
	}

	// Create the combined image
	combinedImg := image.NewRGBA(image.Rect(0, 0, totalWidth, maxHeight))
	offsetX := 0
	for _, img := range images {
		rect := img.Bounds()
		draw.Draw(combinedImg, image.Rect(offsetX, 0, offsetX+rect.Dx(), rect.Dy()), img, rect.Min, draw.Src)
		offsetX += rect.Dx()
	}

	return combinedImg
}

// Save PNG image
func savePNGImage(img *image.RGBA, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file %s: %v", filename, err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}
}
func contains(arr []int, val int) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

// Function to calculate the maximum value from multiple datasets
func calculateMax(data ...[]float64) float64 {
	max := 0.0
	for _, dataset := range data {
		for _, value := range dataset {
			if value > max {
				max = value
			}
		}
	}
	return max
}
func LogValueFormatter(v interface{}) string {
	if value, ok := v.(float64); ok && value > 0 {
		return fmt.Sprintf("%.2f", math.Log10(value))
	}
	return "0"
}
func clampValues(data []float64, min, max float64) []float64 {
	clamped := make([]float64, len(data))
	for i, v := range data {
		if v < min {
			clamped[i] = min
		} else if v > max {
			clamped[i] = max
		} else {
			clamped[i] = v
		}
	}
	return clamped
}

// Modified function definition
func createInfectionGraph(frameNum int, virionOnly, dipOnly, both []float64, showLegend bool) *image.RGBA {
	graphWidth := GRID_SIZE * CELL_SIZE * 2
	graphHeight := 200

	if frameNum < 1 {
		log.Fatalf("Not enough data to render the graph: frameNum = %d", frameNum)
	}

	virionOnly = clampValues(virionOnly, 0.00, yMax)
	dipOnly = clampValues(dipOnly, 0.00, yMax)
	both = clampValues(both, 0.00, yMax)

	// Dynamically set legend name
	var series []chart.Series

	series = []chart.Series{
		chart.ContinuousSeries{
			Name:    "Infected by Virion Only",
			XValues: createTimeSeries(frameNum),
			YValues: virionOnly,
			Style:   chart.Style{StrokeColor: chart.ColorRed, StrokeWidth: 6.0},
		},
		chart.ContinuousSeries{
			Name:    "Infected by DIP Only",
			XValues: createTimeSeries(frameNum),
			YValues: dipOnly,
			Style:   chart.Style{StrokeColor: chart.ColorGreen, StrokeWidth: 6.0},
		},
		chart.ContinuousSeries{
			Name:    "Infected by Both",
			XValues: createTimeSeries(frameNum),
			YValues: both,
			Style:   chart.Style{StrokeColor: drawing.Color{R: 255, G: 165, B: 0, A: 255}, StrokeWidth: 8.0},
		},
	}

	graph := chart.Chart{
		Width:  GRID_SIZE * CELL_SIZE * 1.51,
		Height: 100,
		XAxis: chart.XAxis{
			Style: chart.Style{FontSize: 10.0},
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%d", int(v.(float64)))
			},
			Ticks: generateTicks(xMax, ticksInterval),
		},
		YAxis: chart.YAxis{
			Style: chart.Style{FontSize: 10.0},
		},
		Series: series,
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		log.Fatalf("Failed to render graph: %v", err)
	}

	graphImg, _, err := image.Decode(buffer)
	if err != nil {
		log.Fatalf("Failed to decode graph image: %v", err)
	}

	rgbaImg := image.NewRGBA(image.Rect(0, 0, graphWidth, graphHeight))
	draw.Draw(rgbaImg, rgbaImg.Bounds(), graphImg, image.Point{}, draw.Src)

	return rgbaImg
}

// saveCurrentGoFile saves the current Go source file into the specified output folder.
// saveCurrentGoFile saves the current Go source file with its original name and a timestamp.
func saveCurrentGoFile(outputFolder string) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("Unable to get current Go file path")
		return
	}

	// Get the base name of the current Go file (without path)
	originalFileName := filepath.Base(currentFile)

	// Generate timestamp
	timestamp := time.Now().Format("20060102_150405") // Format: YYYYMMDD_HHMMSS

	// Target filename: original filename_timestamp.go
	newFileName := fmt.Sprintf("%s_%s.go", originalFileName[:len(originalFileName)-3], timestamp)
	outputFilePath := filepath.Join(outputFolder, newFileName)
	// Read Go file content
	content, err := ioutil.ReadFile(currentFile)
	if err != nil {
		log.Printf("cant read file %s: %v\n", currentFile, err)
		return
	}

	// Ensure target folder exists
	if err := os.MkdirAll(outputFolder, os.ModePerm); err != nil {
		log.Printf("cant make outputfolder %s: %v\n", outputFolder, err)
		return
	}

	// Write file
	err = ioutil.WriteFile(outputFilePath, content, 0644)
	if err != nil {
		log.Printf("cant save file to %s: %v\n", outputFilePath, err)
		return
	}

	log.Printf("file successfully saved in %s\n", outputFilePath)
}

func getNextFolderNumber(basePath string) int {
	files, err := os.ReadDir(basePath)
	if err != nil {
		log.Fatalf("Failed to read directory %s: %v", basePath, err)
	}

	maxNumber := 0
	for _, file := range files {
		if file.IsDir() {
			// Try to parse number from folder name
			var folderNumber int
			_, err := fmt.Sscanf(file.Name(), "%d", &folderNumber)
			if err == nil && folderNumber > maxNumber {
				maxNumber = folderNumber
			}
		}
	}
	return maxNumber + 1 // Return next available number
}

func transformToLogScale(data []float64) []float64 {
	transformed := make([]float64, len(data))
	for i, value := range data {
		if value > 0 {
			transformed[i] = math.Log10(value)
		} else {
			transformed[i] = math.Log10(0.0001) // handle log(0) case, use a very small value instead
		}
	}
	return transformed
}

func createTimeSeries(frameNum int) []float64 {
	if frameNum < 1 {
		return []float64{0, 1} // Return a default time series if not enough data
	}

	timeSeries := make([]float64, frameNum+1)
	for i := 0; i <= frameNum; i++ {
		timeSeries[i] = float64(i)
	}
	return timeSeries
}

// Function to calculate total virions in the grid
func (g *Grid) totalVirions() int {
	totalVirions := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			totalVirions += g.localVirions[i][j]
		}
	}
	return totalVirions
}

// Function to calculate total DIPs in the grid
func (g *Grid) totalDIPs() int {
	totalDIPs := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			totalDIPs += g.localDips[i][j]
		}
	}
	return totalDIPs
}

// Function to calculate the total number of regrowth cells in the grid
func (g *Grid) calculateRegrowthCount() int {
	regrowthCells := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == REGROWTH {
				regrowthCells++
				g.timeSinceRegrowth[i][j] += TIMESTEP
			}
		}
	}
	return regrowthCells
}

// Function to calculate the percentage of susceptible cells in the grid
func (g *Grid) calculateSusceptiblePercentage() float64 {
	totalCells := GRID_SIZE * GRID_SIZE
	susceptibleCells := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == SUSCEPTIBLE {
				susceptibleCells++
				g.timeSinceSusceptible[i][j] += TIMESTEP
			}
		}
	}
	return (float64(susceptibleCells) / float64(totalCells)) * 100
}

// Function to calculate the percentage of regrowthed or antiviral cells
func (g *Grid) calculateRegrowthedOrAntiviralPercentage() float64 {
	totalCells := GRID_SIZE * GRID_SIZE
	regrowthedOrAntiviralCells := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == REGROWTH || g.state[i][j] == ANTIVIRAL {
				regrowthedOrAntiviralCells++
			}
		}
	}
	return (float64(regrowthedOrAntiviralCells) / float64(totalCells)) * 100
}

// Function to calculate the percentage of infected cells (both virion and DIP infections)
func (g *Grid) calculateInfectedPercentage() float64 {
	totalCells := GRID_SIZE * GRID_SIZE
	infectedCells := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == INFECTED_VIRION || g.state[i][j] == INFECTED_DIP || g.state[i][j] == INFECTED_BOTH {
				infectedCells++
			}
		}
	}
	return (float64(infectedCells) / float64(totalCells)) * 100
}

// Function to calculate the percentage of DIP-only infected cells
func (g *Grid) calculateInfectedDIPOnlyPercentage() float64 {
	totalCells := GRID_SIZE * GRID_SIZE
	infectedDIPOnlyCells := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == INFECTED_DIP {
				infectedDIPOnlyCells++
			}
		}
	}
	return (float64(infectedDIPOnlyCells) / float64(totalCells)) * 100
}

// Function to calculate the percentage of cells infected by both virions and DIPs
func (g *Grid) calculateInfectedBothPercentage() float64 {
	totalCells := GRID_SIZE * GRID_SIZE
	infectedBothCells := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == INFECTED_BOTH {
				infectedBothCells++
			}
		}
	}
	return (float64(infectedBothCells) / float64(totalCells)) * 100
}

// Function to calculate the percentage of antiviral cells (if antiviral state is modeled)
func (g *Grid) calculateAntiviralPercentage() float64 {
	totalCells := GRID_SIZE * GRID_SIZE
	antiviralCells := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == ANTIVIRAL {
				antiviralCells++
			}
		}
	}
	return (float64(antiviralCells) / float64(totalCells)) * 100
}

// Function to calculate the percentage of uninfected cells (susceptible and regrowth cells)
func (g *Grid) calculateUninfectedPercentage() float64 {
	totalCells := GRID_SIZE * GRID_SIZE
	uninfectedCells := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == SUSCEPTIBLE || g.state[i][j] == REGROWTH {
				uninfectedCells++
			}
		}
	}
	return (float64(uninfectedCells) / float64(totalCells)) * 100
}

// Function to calculate plaque percentage (for simplicity, counting dead cells as plaques)
func (g *Grid) calculatePlaquePercentage() float64 {
	totalCells := GRID_SIZE * GRID_SIZE
	plaqueCells := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == DEAD {
				plaqueCells++
			}
		}
	}
	return (float64(plaqueCells) / float64(totalCells)) * 100
}

// Function to calculate the percentage of dead cells
func calculateDeadCellPercentage(grid [GRID_SIZE][GRID_SIZE]int) float64 {
	totalCells := GRID_SIZE * GRID_SIZE
	deadCells := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if grid[i][j] == DEAD {
				deadCells++
			}
		}
	}
	return (float64(deadCells) / float64(totalCells)) * 100
}

// Function to calculate the number of cells infected by virion only
func (g *Grid) calculateVirionOnlyInfected() int {
	virionOnlyInfected := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == INFECTED_VIRION {
				virionOnlyInfected++
			}
		}
	}
	return virionOnlyInfected
}

// Function to calculate the number of cells infected by DIP only
func (g *Grid) calculateDipOnlyInfected() int {
	dipOnlyInfected := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == INFECTED_DIP {
				dipOnlyInfected++
			}
		}
	}
	return dipOnlyInfected
}

// Function to calculate the number of cells infected by both virion and DIP
func (g *Grid) calculateBothInfected() int {
	bothInfected := 0
	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			if g.state[i][j] == INFECTED_BOTH {
				bothInfected++
			}
		}
	}
	return bothInfected
}

var precomputedRing [][2]int

func precomputeRing(radius int) [][2]int {
	var offsets [][2]int
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			if dx*dx+dy*dy <= radius*radius {
				offsets = append(offsets, [2]int{dx, dy})
			}
		}
	}
	rand.Shuffle(len(offsets), func(i, j int) { offsets[i], offsets[j] = offsets[j], offsets[i] })
	return offsets
}

func precomputeIFNArea(radius int) [][2]int {
	var area [][2]int
	for di := -radius; di <= radius; di++ {
		for dj := -radius; dj <= radius; dj++ {
			distance := math.Sqrt(float64(di*di + dj*dj))
			// Include only cells within the radius
			if distance <= float64(radius) {
				area = append(area, [2]int{di, dj})
			}
		}
	}
	return area
}

// Add this new function, based on the competition mechanism from the paper

// Calculate neighbor relationships
func (g *Grid) initializeNeighbors() {

	precomputedRingV := precomputeRing(jumpRadiusV)
	precomputedRingD := precomputeRing(jumpRadiusD)

	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {
			// Initialize the virion neighbors based on jumpRadiusV

			///////////////////////////////////////////
			indexV := 0
			for _, offset := range precomputedRingV {
				newI, newJ := i+offset[0], j+offset[1]
				// Ensure the new indices are within bounds
				if newI >= 0 && newI < GRID_SIZE && newJ >= 0 && newJ < GRID_SIZE {
					g.neighborsRingVirion[i][j][indexV] = [2]int{newI, newJ}
					indexV++
				}

				// Stop if we have filled all available spots
				if indexV >= len(g.neighborsRingVirion[i][j]) {
					break
				}
			}
			for ; indexV < len(g.neighborsRingVirion[i][j]); indexV++ {
				g.neighborsRingVirion[i][j][indexV] = [2]int{-1, -1}
			}

			// Initialize the DIP neighbors based on jumpRadiusD
			indexD := 0
			for _, offset := range precomputedRingD {
				newI, newJ := i+offset[0], j+offset[1]
				// Ensure the new indices are within bounds
				if newI >= 0 && newI < GRID_SIZE && newJ >= 0 && newJ < GRID_SIZE {
					g.neighborsRingDIP[i][j][indexD] = [2]int{newI, newJ}
					indexD++
				}
				// Stop if we have filled all available spots
				if indexD >= len(g.neighborsRingDIP[i][j]) {
					break
				}
			}

			for ; indexD < len(g.neighborsRingDIP[i][j]); indexD++ {
				g.neighborsRingDIP[i][j][indexD] = [2]int{-1, -1}
			}

			if ifnWave == true {

				precomputedIFNArea := precomputeIFNArea(IFN_wave_radius)
				// Initialize neighbors for IFN area
				var ifnAreaNeighbors [][2]int
				for _, offset := range precomputedIFNArea {
					newI, newJ := i+offset[0], j+offset[1]
					// Ensure the new indices are within grid bounds
					if newI >= 0 && newI < GRID_SIZE && newJ >= 0 && newJ < GRID_SIZE {
						ifnAreaNeighbors = append(ifnAreaNeighbors, [2]int{newI, newJ})
					}
				}
				g.neighborsIFNArea[i][j] = ifnAreaNeighbors

			}

		}

	}

	invalidNeighbor := [2]int{-1, -1} // Invalid neighbor coordinate

	for i := 0; i < GRID_SIZE; i++ {
		for j := 0; j < GRID_SIZE; j++ {

			if i%2 == 0 && j%2 == 0 {
				// Even centerX, even centerY
				// Neighbors at distance 1
				g.neighbors1[i][j] = [6][2]int{
					{i - 1, j},     // left up
					{i + 1, j},     // right up
					{i, j - 1},     // up
					{i, j + 1},     // down
					{i - 1, j + 1}, // left down
					{i + 1, j + 1}, // right down
				}
				// Neighbors at distance 2
				g.neighbors2[i][j] = [6][2]int{
					{i, j - 2},     // up
					{i, j + 2},     // down
					{i - 2, j - 1}, // left up
					{i + 2, j - 1}, // right up
					{i - 2, j + 1}, // left down
					{i + 2, j + 1}, // right down
				}
				// Neighbors at distance 3
				g.neighbors3[i][j] = [6][2]int{
					{i - 2, j},     // left
					{i + 2, j},     // right
					{i - 1, j - 1}, // up left
					{i + 1, j - 1}, // up right
					{i - 1, j - 2}, // up left
					{i + 1, j - 2}, // up right
				}
			} else if i%2 == 1 && j%2 == 0 {
				// Odd centerX, even centerY
				g.neighbors1[i][j] = [6][2]int{
					{i - 1, j},     // left up
					{i + 1, j},     // right up
					{i, j - 1},     // up
					{i, j + 1},     // down
					{i - 1, j + 1}, // left down
					{i + 1, j + 1}, // right down
				}
				g.neighbors2[i][j] = [6][2]int{
					{i, j - 2},     // up
					{i, j + 2},     // down
					{i - 2, j - 1}, // left up
					{i + 2, j - 1}, // right up
					{i - 2, j + 1}, // left down
					{i + 2, j + 1}, // right down
				}
				g.neighbors3[i][j] = [6][2]int{
					{i - 2, j},     // left
					{i + 2, j},     // right
					{i - 1, j - 1}, // up left
					{i + 1, j - 1}, // up right
					{i - 1, j + 2}, // down left
					{i + 1, j + 2}, // down right
				}
			} else if i%2 == 0 && j%2 == 1 {
				// Even centerX, odd centerY
				g.neighbors1[i][j] = [6][2]int{
					{i - 1, j},     // left up
					{i + 1, j},     // right up
					{i, j - 1},     // up
					{i, j + 1},     // down
					{i - 1, j + 1}, // left down
					{i + 1, j + 1}, // right down
				}
				g.neighbors2[i][j] = [6][2]int{
					{i, j - 2},     // up
					{i, j + 2},     // down
					{i - 2, j - 1}, // left up
					{i + 2, j - 1}, // right up
					{i - 2, j + 1}, // left down
					{i + 2, j + 1}, // right down
				}
				g.neighbors3[i][j] = [6][2]int{
					{i - 2, j},     // left
					{i + 2, j},     // right
					{i - 1, j - 1}, // up left
					{i + 1, j - 1}, // up right
					{i - 1, j - 2}, // down left
					{i + 1, j - 2}, // down right
				}
			} else if i%2 == 1 && j%2 == 1 {
				// Odd centerX, odd centerY
				g.neighbors1[i][j] = [6][2]int{
					{i - 1, j}, {i + 1, j}, {i, j - 1}, {i, j + 1}, {i - 1, j + 1}, {i + 1, j + 1},
				}
				g.neighbors2[i][j] = [6][2]int{
					{i, j - 2}, {i, j + 2}, {i - 2, j - 1}, {i + 2, j - 1}, {i - 2, j + 1}, {i + 2, j + 1},
				}
				g.neighbors3[i][j] = [6][2]int{
					{i - 2, j}, {i + 2, j}, {i - 1, j - 1}, {i + 1, j - 1}, {i - 1, j + 2}, {i + 1, j + 2},
				}
			}

			// Remove neighbors that are out of bounds by setting them to invalid values
			for n := 0; n < 6; n++ {
				if g.neighbors1[i][j][n][0] < 0 || g.neighbors1[i][j][n][0] >= GRID_SIZE || g.neighbors1[i][j][n][1] < 0 || g.neighbors1[i][j][n][1] >= GRID_SIZE {
					g.neighbors1[i][j][n] = invalidNeighbor
				}
				if g.neighbors2[i][j][n][0] < 0 || g.neighbors2[i][j][n][0] >= GRID_SIZE || g.neighbors2[i][j][n][1] < 0 || g.neighbors2[i][j][n][1] >= GRID_SIZE {
					g.neighbors2[i][j][n] = invalidNeighbor
				}
				if g.neighbors3[i][j][n][0] < 0 || g.neighbors3[i][j][n][0] >= GRID_SIZE || g.neighbors3[i][j][n][1] < 0 || g.neighbors3[i][j][n][1] >= GRID_SIZE {
					g.neighbors3[i][j][n] = invalidNeighbor
				}
			}
		}

	}

	fmt.Println("Neighbors initialized")

}

// Update the state of the grid at each time step
func (g *Grid) update(frameNum int) {
	newGrid := g.state

	if ifnWave == true {
		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				g.stateChanged[i][j] = false

			}
		}

		// Step 3: Update max global IFN if needed
		if globalIFN < 0 {
			globalIFN = -1.0
		}
		if globalIFN > maxGlobalIFN {

			maxGlobalIFN = globalIFN

		}
		fmt.Printf("Global IFN concentration: %.2f\n", globalIFN)

		// Traverse the grid
		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				// Only consider cells that are in the SUSCEPTIBLE or REGROWTH state

				var regional_sumIFN float64
				neighborsCount := len(g.neighborsIFNArea[i][j])

				if ifn_half_life != 0 {
					for i := 0; i < GRID_SIZE; i++ {
						for j := 0; j < GRID_SIZE; j++ {
							// Update IFN amount using half-life formula
							factorIFN := math.Pow(0.5, float64(TIMESTEP)/ifn_half_life)
							g.IFNConcentration[i][j] *= factorIFN
							// Remove IFN if concentration is below threshold
							if g.IFNConcentration[i][j] < (1.0 / (float64(GRID_SIZE) * float64(GRID_SIZE))) {
								g.IFNConcentration[i][j] = 0
							}
						}
					}
				}

				// Sum the IFN concentration within the IFN area
				for _, neighbor := range g.neighborsIFNArea[i][j] {
					ni, nj := neighbor[0], neighbor[1]

					regional_sumIFN += g.IFNConcentration[ni][nj]
				}

				// Calculate the average IFN concentration if there are neighbors within the radius
				var regionalAverageIFN float64
				if neighborsCount > 0 {
					regionalAverageIFN = regional_sumIFN / float64(neighborsCount)
				} else {
					regionalAverageIFN = 0 // Default to 0 if no neighbors, though this should rarely occur
				}

				if g.state[i][j] == SUSCEPTIBLE || g.state[i][j] == REGROWTH || g.state[i][j] == INFECTED_DIP {
					if g.IFNConcentration[i][j] > 0 && TAU > 0 {

						if g.antiviralDuration[i][j] <= -1 {
							g.antiviralDuration[i][j] = int(rand.NormFloat64()*float64(TAU)/4 + float64(TAU))
							g.timeSinceAntiviral[i][j] = 0
						} else if g.timeSinceAntiviral[i][j] <= int(g.antiviralDuration[i][j]) {
							g.timeSinceAntiviral[i][j] += TIMESTEP
						} else {

							g.previousStates[i][j] = g.state[i][j]
							newGrid[i][j] = ANTIVIRAL

							g.timeSinceAntiviral[i][j] = -2
							g.totalAntiviralTime += g.antiviralDuration[i][j]
							if g.state[i][j] == ANTIVIRAL && !g.antiviralFlag[i][j] {
								g.antiviralFlag[i][j] = true
								g.antiviralCellCount++
							}

						}
					}

					if g.state[i][j] == SUSCEPTIBLE || g.state[i][j] == REGROWTH {
						// Check if the cell is infected by virions or DIPs
						if g.localVirions[i][j] > 0 || g.localDips[i][j] > 0 {
							// Calculate the infection probabilities
							if R == 0 || TAU == 0 {
								perParticleInfectionChance_V = RHO
							} else if VStimulateIFN == true && R > 0 { // R=1
								perParticleInfectionChance_V = RHO * math.Exp(-ALPHA*(regionalAverageIFN/float64(R)))
							} else if !VStimulateIFN { // usually only DIP stimulate IFN in this situlation
								perParticleInfectionChance_V = RHO * math.Exp(-ALPHA*(regionalAverageIFN))
							}
							var probabilityVInfection, probabilityDInfection float64

							// Virion infection probability
							probabilityVInfection = 1 - math.Pow(1-perParticleInfectionChance_V, float64(g.localVirions[i][j]))
							infectedByVirion := rand.Float64() <= probabilityVInfection

							// DIP infection probability
							probabilityDInfection = 1 - math.Pow(1-(RHO*math.Exp(-ALPHA*(regionalAverageIFN))), float64(g.localDips[i][j]))
							infectedByDip := rand.Float64() <= probabilityDInfection

							// Determine the infection state based on virion and DIP infection
							if infectedByVirion && infectedByDip {
								newGrid[i][j] = INFECTED_BOTH
								g.timeSinceSusceptible[i][j] = -1
								g.timeSinceRegrowth[i][j] = -1
							} else if infectedByVirion {
								newGrid[i][j] = INFECTED_VIRION
								g.timeSinceSusceptible[i][j] = -1
								g.timeSinceRegrowth[i][j] = -1
							} else if infectedByDip {
								newGrid[i][j] = INFECTED_DIP
								g.timeSinceSusceptible[i][j] = -1
								g.timeSinceRegrowth[i][j] = -1
							}
						}

						// Mark the state as changed if the cell is infected
						if newGrid[i][j] != g.state[i][j] {
							g.stateChanged[i][j] = true
						}
					}

				}

			}
		}

		// Process infected cells
		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {

				var regional_sumIFN float64

				// Sum the IFN concentration within the IFN area
				for _, neighbor := range g.neighborsIFNArea[i][j] {
					ni, nj := neighbor[0], neighbor[1]
					regional_sumIFN += g.IFNConcentration[ni][nj]
				}

				// Note: regionalAverageIFN is not used in this section

				if g.state[i][j] == INFECTED_VIRION || g.state[i][j] == INFECTED_DIP || g.state[i][j] == INFECTED_BOTH {

					// update infected by V or BOTH cells become dead
					if g.state[i][j] == INFECTED_VIRION || g.state[i][j] == INFECTED_BOTH {
						if g.lysisThreshold[i][j] == -1 {
							g.lysisThreshold[i][j] = int(rand.NormFloat64()*STANDARD_LYSIS_TIME + MEAN_LYSIS_TIME)
						}
						g.timeSinceInfectVorBoth[i][j] += TIMESTEP
						g.timeSinceInfectDIP[i][j] = -1

						// Check if the cell should lyse and release virions and DIPs
						if g.lysisThreshold[i][j] > 0 && g.timeSinceInfectVorBoth[i][j] >= g.lysisThreshold[i][j] {

							// After lysis, the cell becomes DEAD and virions and DIPs are spread to neighbors
							if g.state[i][j] == INFECTED_VIRION {
								totalDeadFromV++ // Increase INFECTED_VIRION death count
							} else if g.state[i][j] == INFECTED_BOTH {
								totalDeadFromBoth++ // Increase INFECTED_BOTH death count
							}

							newGrid[i][j] = DEAD
							g.state[i][j] = DEAD
							g.timeSinceDead[i][j] = 0
							g.timeSinceInfectVorBoth[i][j] = -1
							g.timeSinceInfectDIP[i][j] = -1
							g.lysisThreshold[i][j] = -1

							////////// work on percentage of jump randomly and percentage of spread cell to cell without jump
							///////////// for k_jumpR percent cells that jump reandomly
							if par_celltocell_random == true {
								// Calculate adjusted burst size for DIPs based on local ratio
								totalVirionsAtCell := g.localVirions[i][j]
								totalDIPsAtCell := g.localDips[i][j]
								adjustedBurstSizeD := BURST_SIZE_D
								if totalVirionsAtCell > 0 {
									dipVirionRatio := float64(totalDIPsAtCell) / float64(totalVirionsAtCell)
									adjustedBurstSizeD += int(float64(BURST_SIZE_D) * dipVirionRatio)
								}
								//  ---------------------------------------
								// Partition mode: split particles between random jump and cell-to-cell
								randomVirions := int(math.Floor(float64(BURST_SIZE_V) * k_JumpR))
								virionsForLocalDiffusion := BURST_SIZE_V - randomVirions

								randomDIPs := int(math.Floor(float64(adjustedBurstSizeD) * k_JumpR))
								dipsForLocalDiffusion := adjustedBurstSizeD - randomDIPs

								// Handle random jumps
								for v := 0; v < randomVirions; v++ {
									ni, nj := rand.Intn(GRID_SIZE), rand.Intn(GRID_SIZE)
									g.localVirions[ni][nj]++
									g.totalRandomJumpVirions++
								}
								for d := 0; d < randomDIPs; d++ {
									ni, nj := rand.Intn(GRID_SIZE), rand.Intn(GRID_SIZE)
									g.localDips[ni][nj]++
									g.totalRandomJumpDIPs++
								}

								// Handle local diffusion
								// Handle local diffusion with localVirions & localDIPs (keep original logic unchanged)
								if virionsForLocalDiffusion > 0 || dipsForLocalDiffusion > 0 {
									// Calculate the total number of valid neighbors
									totalNeighbors := 0

									// Count valid neighbors from neighbors1
									for _, dir := range g.neighbors1[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}
									// Count valid neighbors from neighbors2
									for _, dir := range g.neighbors2[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}
									// Count valid neighbors from neighbors3
									for _, dir := range g.neighbors3[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}

									if totalNeighbors == 0 {
										return
									}

									// Calculate the distribution based on the ratio √3 : 2√3 : 3
									sqrt3 := math.Sqrt(3)
									ratio1 := 1.0               // sqrt3     // Weight for neighbors1
									ratio2 := 1.0 / 2           // 2 * sqrt3 // Weight for neighbors2
									ratio3 := 1.0 / (3 / sqrt3) // 3.0       // Weight for neighbors3
									totalRatio := ratio1*float64(len(g.neighbors1[i][j])) +
										ratio2*float64(len(g.neighbors2[i][j])) +
										ratio3*float64(len(g.neighbors3[i][j]))

									// Calculate virions for each neighbor group
									virionsForNeighbors1 := int(math.Floor(float64(virionsForLocalDiffusion) * (ratio1 * float64(len(g.neighbors1[i][j]))) / totalRatio))
									virionsForNeighbors2 := int(math.Floor(float64(virionsForLocalDiffusion) * (ratio2 * float64(len(g.neighbors2[i][j]))) / totalRatio))
									virionsForNeighbors3 := int(math.Floor(float64(virionsForLocalDiffusion) * (ratio3 * float64(len(g.neighbors3[i][j]))) / totalRatio))

									// Calculate remaining virions
									remainingVirions := virionsForLocalDiffusion - (virionsForNeighbors1 + virionsForNeighbors2 + virionsForNeighbors3)

									// Distribute remaining virions based on ratio
									for remainingVirions > 0 {
										randVal := rand.Float64() * totalRatio
										if randVal < ratio1 && len(g.neighbors1[i][j]) > 0 {
											virionsForNeighbors1++
										} else if randVal < (ratio1+ratio2) && len(g.neighbors2[i][j]) > 0 {
											virionsForNeighbors2++
										} else if len(g.neighbors3[i][j]) > 0 {
											virionsForNeighbors3++
										}
										remainingVirions--
									}

									// Calculate DIPs for each neighbor group (same logic)
									dipsForNeighbors1 := int(math.Floor(float64(dipsForLocalDiffusion) * (ratio1 * float64(len(g.neighbors1[i][j]))) / totalRatio))
									dipsForNeighbors2 := int(math.Floor(float64(dipsForLocalDiffusion) * (ratio2 * float64(len(g.neighbors2[i][j]))) / totalRatio))
									dipsForNeighbors3 := int(math.Floor(float64(dipsForLocalDiffusion) * (ratio3 * float64(len(g.neighbors3[i][j]))) / totalRatio))

									remainingDIPs := dipsForLocalDiffusion - (dipsForNeighbors1 + dipsForNeighbors2 + dipsForNeighbors3)

									for remainingDIPs > 0 {
										randVal := rand.Float64() * totalRatio
										if randVal < ratio1 && len(g.neighbors1[i][j]) > 0 {
											dipsForNeighbors1++
										} else if randVal < (ratio1+ratio2) && len(g.neighbors2[i][j]) > 0 {
											dipsForNeighbors2++
										} else if len(g.neighbors3[i][j]) > 0 {
											dipsForNeighbors3++
										}
										remainingDIPs--
									}

									// Distribute virions to neighbors1
									for _, dir := range g.neighbors1[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											if g.state[ni][nj] == SUSCEPTIBLE {
												g.localVirions[ni][nj] += virionsForNeighbors1 / len(g.neighbors1[i][j])
												g.localDips[ni][nj] += dipsForNeighbors1 / len(g.neighbors1[i][j])
											}
										}
									}

									// Distribute virions to neighbors2
									for _, dir := range g.neighbors2[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											g.localVirions[ni][nj] += virionsForNeighbors2 / len(g.neighbors2[i][j])
											g.localDips[ni][nj] += dipsForNeighbors2 / len(g.neighbors2[i][j])
										}
									}

									// Distribute virions to neighbors3
									for _, dir := range g.neighbors3[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											g.localVirions[ni][nj] += virionsForNeighbors3 / len(g.neighbors3[i][j])
											g.localDips[ni][nj] += dipsForNeighbors3 / len(g.neighbors3[i][j])
										}
									}
								}

							} else if par_celltocell_random == false {
								//////////////////////////////
								fmt.Println("parition particles jump celltocell and randomly is false")

								if !allowVirionJump && !allowDIPJump {
									fmt.Println("Virion and DIP jump are both disabled, NO JUMP")
									// Calculate the total number of valid neighbors
									totalNeighbors := 0

									// Count valid neighbors from neighbors1
									for _, dir := range g.neighbors1[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}
									// Count valid neighbors from neighbors2
									for _, dir := range g.neighbors2[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}
									// Count valid neighbors from neighbors3
									for _, dir := range g.neighbors3[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}

									// If there are no valid neighbors, return early
									if totalNeighbors == 0 {
										return
									}

									// Calculate the distribution of virions and DIPs to each neighbor based on the ratio √3 : 2√3 : 3
									sqrt3 := math.Sqrt(3)
									ratio1 := 1.0               // sqrt3     // Weight for neighbors1
									ratio2 := 1.0 / 2           // 2 * sqrt3 // Weight for neighbors2
									ratio3 := 1.0 / (3 / sqrt3) // 3.0 // Weight for neighbors3
									totalRatio := ratio1*float64(len(g.neighbors1[i][j])) + ratio2*float64(len(g.neighbors2[i][j])) + ratio3*float64(len(g.neighbors3[i][j]))
									// if infected by virion or infected by both:
									// Calculate the number of virions and DIPs assigned to each type of neighbor
									virionsForNeighbors1 := int(math.Floor(float64(BURST_SIZE_V) * (ratio1 * float64(len(g.neighbors1[i][j]))) / totalRatio))
									virionsForNeighbors2 := int(math.Floor(float64(BURST_SIZE_V) * (ratio2 * float64(len(g.neighbors2[i][j]))) / totalRatio))
									virionsForNeighbors3 := int(math.Floor(float64(BURST_SIZE_V) * (ratio3 * float64(len(g.neighbors3[i][j]))) / totalRatio))

									// Calculate the remaining virions and DIPs
									remainingVirions := BURST_SIZE_V - (virionsForNeighbors1 + virionsForNeighbors2 + virionsForNeighbors3)

									// // Randomly distribute the remaining virions based on the ratio
									for remainingVirions > 0 {
										randVal := rand.Float64() * totalRatio
										if randVal < ratio1 && len(g.neighbors1[i][j]) > 0 {
											virionsForNeighbors1++
										} else if randVal < (ratio1+ratio2) && len(g.neighbors2[i][j]) > 0 {
											virionsForNeighbors2++
										} else if len(g.neighbors3[i][j]) > 0 {
											virionsForNeighbors3++
										}
										remainingVirions--
									}
									// if infected by vrion only or both:

									totalVirionsAtCell := g.localVirions[i][j]
									totalDIPsAtCell := g.localDips[i][j]

									// Ensure we avoid division by zero
									adjustedBurstSizeD := 0
									if totalVirionsAtCell > 0 {
										// Adjust BURST_SIZE_D based on the DIP-to-virion ratio at this cell
										dipVirionRatio := float64(totalDIPsAtCell) / float64(totalVirionsAtCell)
										adjustedBurstSizeD = BURST_SIZE_D + int(math.Floor(float64(BURST_SIZE_D)*dipVirionRatio))
									}

									// Distribute DIPs to neighbors based on the adjusted BURST_SIZE_D
									dipsForNeighbors1 := int(math.Floor(float64(adjustedBurstSizeD) * (ratio1 * float64(len(g.neighbors1[i][j]))) / totalRatio))
									dipsForNeighbors2 := int(math.Floor(float64(adjustedBurstSizeD) * (ratio2 * float64(len(g.neighbors2[i][j]))) / totalRatio))
									dipsForNeighbors3 := int(math.Floor(float64(adjustedBurstSizeD) * (ratio3 * float64(len(g.neighbors3[i][j]))) / totalRatio))
									remainingDips := adjustedBurstSizeD - (dipsForNeighbors1 + dipsForNeighbors2 + dipsForNeighbors3)

									// Randomly distribute the remaining DIPs based on the ratio
									for remainingDips > 0 {
										randVal := rand.Float64() * totalRatio
										if randVal < ratio1 && len(g.neighbors1[i][j]) > 0 {
											dipsForNeighbors1++
										} else if randVal < (ratio1+ratio2) && len(g.neighbors2[i][j]) > 0 {
											dipsForNeighbors2++
										} else if len(g.neighbors3[i][j]) > 0 {
											dipsForNeighbors3++
										}
										remainingDips--
									}
									// Distribute virions and DIPs to neighbors1
									for _, dir := range g.neighbors1[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											if g.state[ni][nj] == SUSCEPTIBLE {
												g.localVirions[ni][nj] += virionsForNeighbors1 / len(g.neighbors1[i][j])
												g.localDips[ni][nj] += dipsForNeighbors1 / len(g.neighbors1[i][j])
											}
										}
									}

									// Distribute virions and DIPs to neighbors2
									for _, dir := range g.neighbors2[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											g.localVirions[ni][nj] += virionsForNeighbors2 / len(g.neighbors2[i][j])
											g.localDips[ni][nj] += dipsForNeighbors2 / len(g.neighbors2[i][j])
										}
									}

									// Distribute virions and DIPs to neighbors3
									for _, dir := range g.neighbors3[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											g.localVirions[ni][nj] += virionsForNeighbors3 / len(g.neighbors3[i][j])
											g.localDips[ni][nj] += dipsForNeighbors3 / len(g.neighbors3[i][j])
										}
									}

								} else { // "Jump" case for either virions, DIPs, or both
									fmt.Println("Virion and DIP jump are allowed to JUMP")
									if allowVirionJump {
										totalVirionsAtCell := g.localVirions[i][j]
										totalDIPsAtCell := g.localDips[i][j]
										adjustedBurstSizeD := 0
										if totalVirionsAtCell > 0 {
											dipVirionRatio := float64(totalDIPsAtCell) / float64(totalVirionsAtCell)
											adjustedBurstSizeD = BURST_SIZE_D + int(math.Floor(float64(BURST_SIZE_D)*dipVirionRatio))
										}
										if jumpRandomly {
											for v := 0; v < BURST_SIZE_V; v++ {
												ni := rand.Intn(GRID_SIZE) // Randomly select a row
												nj := rand.Intn(GRID_SIZE) // Randomly select a column

												// Apply the virion jump
												g.localVirions[ni][nj]++
											}

											// DIP jump randomly to any location
											for d := 0; d < adjustedBurstSizeD; d++ {
												ni := rand.Intn(GRID_SIZE) // Randomly select a row
												nj := rand.Intn(GRID_SIZE) // Randomly select a column

												// Apply the DIP jump
												g.localDips[ni][nj]++
											}
										} else {

											totalVirionsAtCell := g.localVirions[i][j]
											totalDIPsAtCell := g.localDips[i][j]
											adjustedBurstSizeD := 0
											if totalVirionsAtCell > 0 {
												dipVirionRatio := float64(totalDIPsAtCell) / float64(totalVirionsAtCell)
												adjustedBurstSizeD = BURST_SIZE_D + int(math.Floor(float64(BURST_SIZE_D)*dipVirionRatio))
											}

											// Virion jump logic
											virionTargets := make([]int, BURST_SIZE_V)
											for v := 0; v < BURST_SIZE_V; v++ {
												virionTargets[v] = rand.Intn(len(g.neighborsRingVirion[i][j]))
											}

											// Apply virion jumps
											for _, targetIndex := range virionTargets {
												spot := g.neighborsRingVirion[i][j][targetIndex]
												ni, nj := spot[0], spot[1]

												// Ensure the jump target is valid
												if ni < 0 || ni >= GRID_SIZE || nj < 0 || nj >= GRID_SIZE {
													// fmt.Printf("Skipping invalid jump target (%d, %d) from (%d, %d)\n", ni, nj, i, j)
													continue
												}

												// Apply the virion jump
												g.localVirions[ni][nj]++
											}

											// DIP jump logic
											dipTargets := make([]int, adjustedBurstSizeD)
											for d := 0; d < adjustedBurstSizeD; d++ {
												dipTargets[d] = rand.Intn(len(g.neighborsRingDIP[i][j]))
											}

											// Apply DIP jumps
											for _, targetIndex := range dipTargets {
												spot := g.neighborsRingDIP[i][j][targetIndex]
												ni, nj := spot[0], spot[1]

												// Ensure the jump target is valid
												if ni < 0 || ni >= GRID_SIZE || nj < 0 || nj >= GRID_SIZE {
													//fmt.Printf("Skipping invalid jump target (%d, %d) from (%d, %d)\n", ni, nj, i, j)
													continue
												}

												// Apply the DIP jump
												g.localDips[ni][nj]++
											}
										}
									}

									if allowDIPJump {
										totalVirionsAtCell := g.localVirions[i][j]
										totalDIPsAtCell := g.localDips[i][j]
										adjustedBurstSizeD := 0
										if totalVirionsAtCell > 0 {
											dipVirionRatio := float64(totalDIPsAtCell) / float64(totalVirionsAtCell)
											adjustedBurstSizeD = BURST_SIZE_D + int(math.Floor(float64(BURST_SIZE_D)*dipVirionRatio))
										}

										if jumpRandomly {
											go func() {
												for d := 0; d < adjustedBurstSizeD; d++ {
													ni := rand.Intn(GRID_SIZE)
													nj := rand.Intn(GRID_SIZE)
													g.localDips[ni][nj]++
												}
											}()
										} else {
											dipTargets := make([]int, adjustedBurstSizeD)
											for d := 0; d < adjustedBurstSizeD; d++ {
												dipTargets[d] = rand.Intn(len(g.neighborsRingDIP[i][j]))
											}
											go func() {
												for _, targetIndex := range dipTargets {
													spot := g.neighborsRingDIP[i][j][targetIndex]
													ni, nj := spot[0], spot[1]
													if ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
														g.localDips[ni][nj]++
													}
												}
											}()
										}
									}

								}
							}
						}
					}
					// update infected only by DIP or only by virions cells become "infected by both"
					if g.state[i][j] == INFECTED_VIRION || g.state[i][j] == INFECTED_DIP {

						if g.stateChanged[i][j] == false {
							// Check if the cell is infected by virions or DIPs

							if g.localVirions[i][j] > 0 || g.localDips[i][j] > 0 {
								// Calculate the infection probabilities

								if R == 0 || TAU == 0 {
									perParticleInfectionChance_V = RHO
								} else {
									if VStimulateIFN == true { // R=1
										perParticleInfectionChance_V = RHO * math.Exp(-ALPHA*(globalIFNperCell/float64(R)))
									} else if VStimulateIFN == false { // usually only DIP stimulate IFN in this situlation
										perParticleInfectionChance_V = RHO * math.Exp(-ALPHA*(globalIFNperCell))
									}
								}
								var probabilityVInfection, probabilityDInfection float64

								// Virion infection probability
								probabilityVInfection = 1 - math.Pow(1-perParticleInfectionChance_V, float64(g.localVirions[i][j]))
								infectedByVirion := rand.Float64() <= probabilityVInfection

								// DIP infection probability
								probabilityDInfection = 1 - math.Pow(1-(RHO*math.Exp(-ALPHA*(globalIFNperCell))), float64(g.localDips[i][j]))
								infectedByDip := rand.Float64() <= probabilityDInfection

								// Determine the infection state based on virion and DIP infection
								if infectedByVirion && infectedByDip {
									newGrid[i][j] = INFECTED_BOTH
								} else if infectedByVirion {
									newGrid[i][j] = INFECTED_VIRION
								} else if infectedByDip {
									newGrid[i][j] = INFECTED_DIP
								}
							}

						}

						if g.state[i][j] == INFECTED_VIRION || g.state[i][j] == INFECTED_BOTH {

							if g.timeSinceInfectVorBoth[i][j] > IFN_DELAY+int(math.Floor(rand.NormFloat64()*float64(STD_IFN_DELAY))) && TAU > 0 {
								adjusted_DIP_IFN_stimulate := 1.0
								// if g.intraWT[i][j] > 0 {
								// 	dvgWtRatio := float64(g.intraDVG[i][j]) / float64(g.intraWT[i][j])
								// 	if g.state[i][j] == INFECTED_BOTH {
								// 		adjusted_DIP_IFN_stimulate *= dvgWtRatio * BOTH_IFN_stimulate_ratio
								// 	}
								// }
								adjusted_DIP_IFN_stimulate = BOTH_IFN_stimulate_ratio
								var totalIncreaseAmount float64
								if VStimulateIFN == true {

									if g.state[i][j] == INFECTED_VIRION {

										totalIncreaseAmount = float64(R) * float64(TIMESTEP) * ifnBothFold
									} else if g.state[i][j] == INFECTED_BOTH {
										totalIncreaseAmount = (float64(R) + adjusted_DIP_IFN_stimulate) * float64(TIMESTEP)
									}
								} else if VStimulateIFN == false {

									if g.state[i][j] == INFECTED_VIRION {

										totalIncreaseAmount = 0.0
									} else if g.state[i][j] == INFECTED_BOTH {
										totalIncreaseAmount = (adjusted_DIP_IFN_stimulate) * float64(TIMESTEP)
									}
									fmt.Println("totalIncreaseAmount", totalIncreaseAmount)
								}

								cellCount := len(g.neighborsIFNArea[i][j])

								if cellCount > 0 {
									averageIncreaseAmount := totalIncreaseAmount / float64(cellCount)

									for _, offset := range g.neighborsIFNArea[i][j] {
										ni, nj := offset[0], offset[1]

										if ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											g.IFNConcentration[ni][nj] += averageIncreaseAmount

											g.IFNConcentration[ni][nj] += averageIncreaseAmount

											globalIFN += averageIncreaseAmount

										}
									}
								}
							}

						}

						if g.state[i][j] == INFECTED_DIP {
							g.timeSinceInfectDIP[i][j] += TIMESTEP

							if g.timeSinceInfectDIP[i][j] > IFN_DELAY+int(math.Floor(rand.NormFloat64()*float64(STD_IFN_DELAY))) && TAU > 0 {
								// adjusted_DIP_IFN_stimulate := float64(g.intraDVG[i][j]) * D_only_IFN_stimulate_ratio
								adjusted_DIP_IFN_stimulate := D_only_IFN_stimulate_ratio
								totalIncreaseAmount := adjusted_DIP_IFN_stimulate * float64(TIMESTEP)

								cellCount := len(g.neighborsIFNArea[i][j])
								if cellCount > 0 {
									averageIncreaseAmount := totalIncreaseAmount / float64(cellCount)
									for _, offset := range g.neighborsIFNArea[i][j] {
										ni, nj := offset[0], offset[1]

										if ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											g.IFNConcentration[ni][nj] += averageIncreaseAmount
											globalIFN += averageIncreaseAmount
										}
									}
								}
							}
						}

					}

				}
			}
		}
		// Handle potentially regrowing dead cells
		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				if g.state[i][j] == DEAD {
					g.timeSinceDead[i][j] += TIMESTEP

					// Check if any neighboring cells are susceptible, allowing for regrowth
					canRegrow := false
					neighbors := g.neighbors1[i][j]

					// Iterate over the neighbors and check if any are SUSCEPTIBLE
					for _, neighbor := range neighbors {
						ni, nj := neighbor[0], neighbor[1]

						// Ensure the neighbor indices are valid (within grid bounds)
						if ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {

							//if g.timeSinceSusceptible[ni][nj]+g.timeSinceAntiviral[ni][nj] > int(math.Floor(rand.NormFloat64()*REGROWTH_STD+REGROWTH_MEAN)) || g.timeSinceRegrowth[ni][nj]+g.timeSinceAntiviral[ni][nj] > int(math.Floor(rand.NormFloat64()*REGROWTH_STD+REGROWTH_MEAN)) {
							//	canRegrow = true
							//  break
							//}
							if g.state[ni][nj] == SUSCEPTIBLE || g.state[ni][nj] == ANTIVIRAL {
								canRegrow = true
								break

							}
						}
					}

					// If the conditions are met, the cell regrows
					if canRegrow && g.timeSinceDead[i][j] >= int(rand.NormFloat64()*REGROWTH_STD+REGROWTH_MEAN) {
						newGrid[i][j] = REGROWTH
						g.timeSinceRegrowth[i][j] = 0
						g.timeSinceDead[i][j] = -1

					}

				}
			}
		}

		// Apply the updated grid state
		g.state = newGrid

		// Calculate and log the total virions and DIPs for each time step
		totalVirions, totalDIPs := g.totalVirions(), g.totalDIPs()
		fmt.Printf("Time step %d: Total Virions = %d, Total DIPs = %d\n", frameNum, totalVirions, totalDIPs)

		// Additional calculations based on simulation parameters for tracking purposes
		regrowthCount := g.calculateRegrowthCount()
		susceptiblePercentage := g.calculateSusceptiblePercentage()

		regrowthedOrAntiviralPercentage := g.calculateRegrowthedOrAntiviralPercentage()
		infectedPercentage := g.calculateInfectedPercentage()
		infectedDIPOnlyPercentage := g.calculateInfectedDIPOnlyPercentage()
		infectedBothPercentage := g.calculateInfectedBothPercentage()
		antiviralPercentage := g.calculateAntiviralPercentage()
		deadCellPercentage := calculateDeadCellPercentage(g.state)
		uninfectedPercentage := g.calculateUninfectedPercentage()
		plaquePercentage := g.calculatePlaquePercentage()

		// Log additional data as necessary
		fmt.Printf("Regrowth Count: %d, Susceptible: %.2f%%\n", regrowthCount, susceptiblePercentage)
		fmt.Printf("Regrowthed or Antiviral: %.2f%%, Infected: %.2f%%, DIP Only: %.2f%%, Both Infected: %.2f%%, Antiviral: %.2f%%\n",
			regrowthedOrAntiviralPercentage, infectedPercentage, infectedDIPOnlyPercentage, infectedBothPercentage, antiviralPercentage)
		fmt.Printf("Dead: %.2f%%, Uninfected: %.2f%%, Plaque: %.2f%%\n", deadCellPercentage, uninfectedPercentage, plaquePercentage)

		/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	} else if ifnWave == false { // ifnWave == false

		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				g.stateChanged[i][j] = false
				g.IFNConcentration[i][j] = globalIFN / float64(GRID_SIZE*GRID_SIZE)
			}
		}
		if globalIFN < 0 {
			globalIFN = -1.0
		}
		// Step 3: Update max global IFN if needed
		if globalIFN > maxGlobalIFN {

			maxGlobalIFN = globalIFN

		}

		// Traverse the grid
		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				// Only consider cells that are in the SUSCEPTIBLE or REGROWTH state

				if g.state[i][j] == SUSCEPTIBLE || g.state[i][j] == REGROWTH || g.state[i][j] == INFECTED_DIP {
					if g.IFNConcentration[i][j] > 0 && TAU > 0 {

						if g.antiviralDuration[i][j] == -1 {
							g.antiviralDuration[i][j] = int(math.Floor(rand.NormFloat64()*float64(TAU)/4 + float64(TAU)))
							g.timeSinceAntiviral[i][j] = 0
						} else if g.timeSinceAntiviral[i][j] <= int(g.antiviralDuration[i][j]) {
							g.timeSinceAntiviral[i][j] += TIMESTEP
						} else {

							g.previousStates[i][j] = g.state[i][j]
							newGrid[i][j] = ANTIVIRAL
							g.timeSinceAntiviral[i][j] = -2
							g.totalAntiviralTime += g.antiviralDuration[i][j]
							if g.state[i][j] == ANTIVIRAL && !g.antiviralFlag[i][j] {
								g.antiviralFlag[i][j] = true
								g.antiviralCellCount++
							}

						}

					}

					if g.state[i][j] == SUSCEPTIBLE || g.state[i][j] == REGROWTH {
						// Check if the cell is infected by virions or DIPs
						if g.localVirions[i][j] > 0 || g.localDips[i][j] > 0 {
							// Calculate the infection probabilities

							if R == 0 || TAU == 0 {
								perParticleInfectionChance_V = RHO

							} else {
								if VStimulateIFN == true { // R=1
									perParticleInfectionChance_V = RHO * math.Exp(-ALPHA*(globalIFNperCell/float64(R)))
								} else if VStimulateIFN == false { // usually only DIP stimulate IFN in this situlation
									perParticleInfectionChance_V = RHO * math.Exp(-ALPHA*(globalIFNperCell))
								}
							}

							var probabilityVInfection, probabilityDInfection float64

							// Virion infection probability
							probabilityVInfection = 1 - math.Pow(1-perParticleInfectionChance_V, float64(g.localVirions[i][j]))
							infectedByVirion := rand.Float64() <= probabilityVInfection

							// DIP infection probability
							probabilityDInfection = 1 - math.Pow(1-(RHO*math.Exp(-ALPHA*(globalIFNperCell))), float64(g.localDips[i][j]))
							infectedByDip := rand.Float64() <= probabilityDInfection

							// Determine the infection state based on virion and DIP infection
							if infectedByVirion && infectedByDip {
								newGrid[i][j] = INFECTED_BOTH
								g.timeSinceSusceptible[i][j] = -1
								g.timeSinceRegrowth[i][j] = -1
							} else if infectedByVirion {
								newGrid[i][j] = INFECTED_VIRION
								g.timeSinceSusceptible[i][j] = -1
								g.timeSinceRegrowth[i][j] = -1
							} else if infectedByDip {
								newGrid[i][j] = INFECTED_DIP
								g.timeSinceSusceptible[i][j] = -1
								g.timeSinceRegrowth[i][j] = -1
							}
						}

						// Mark the state as changed if the cell is infected
						if newGrid[i][j] != g.state[i][j] {
							g.stateChanged[i][j] = true
						}
					}

				}

			}
		}

		// Process infected cells, no ifn wave, globally constant ifn
		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				if par_celltocell_random == true {

					allowRandomly := make([][]bool, GRID_SIZE)
					for i := range allowRandomly {
						allowRandomly[i] = make([]bool, GRID_SIZE)
					}

					// Calculate total number of cells allowed for random jumping based on k_JumpR
					totalCells := GRID_SIZE * GRID_SIZE

					randomJumpCells := int(math.Floor(float64(totalCells) * k_JumpR))

					// Randomly select randomJumpCells cells and mark them as allowRandomly
					selectedCells := make(map[[2]int]bool)
					for len(selectedCells) < randomJumpCells {
						ni := rand.Intn(GRID_SIZE)
						nj := rand.Intn(GRID_SIZE)
						selectedCells[[2]int{ni, nj}] = true
					}
					for pos := range selectedCells {
						allowRandomly[pos[0]][pos[1]] = true
					}

				}

				if g.state[i][j] == INFECTED_VIRION || g.state[i][j] == INFECTED_DIP || g.state[i][j] == INFECTED_BOTH {

					// update infected by V or BOTH cells become dead
					if g.state[i][j] == INFECTED_VIRION || g.state[i][j] == INFECTED_BOTH {

						if g.lysisThreshold[i][j] == -1 {
							g.lysisThreshold[i][j] = int(rand.NormFloat64()*STANDARD_LYSIS_TIME + MEAN_LYSIS_TIME)
						}
						g.timeSinceInfectVorBoth[i][j] += TIMESTEP
						g.timeSinceInfectDIP[i][j] = -1

						// Check if the cell should lyse and release virions and DIPs
						if g.timeSinceInfectVorBoth[i][j] > g.lysisThreshold[i][j] {
							if g.state[i][j] == INFECTED_VIRION {
								totalDeadFromV++ // Increase INFECTED_VIRION death count
							} else if g.state[i][j] == INFECTED_BOTH {
								totalDeadFromBoth++ // Increase INFECTED_BOTH death count
							}

							// After lysis, the cell becomes DEAD and virions and DIPs are spread to neighbors
							newGrid[i][j] = DEAD
							g.state[i][j] = DEAD
							g.timeSinceDead[i][j] = 0
							g.timeSinceInfectVorBoth[i][j] = -1
							g.timeSinceInfectDIP[i][j] = -1
							g.lysisThreshold[i][j] = -1

							if par_celltocell_random == true {
								// Calculate adjusted burst size for DIPs based on local ratio
								totalVirionsAtCell := g.localVirions[i][j]
								totalDIPsAtCell := g.localDips[i][j]
								adjustedBurstSizeD := 0
								if totalVirionsAtCell > 0 {
									dipVirionRatio := float64(totalDIPsAtCell) / float64(totalVirionsAtCell)
									adjustedBurstSizeD = BURST_SIZE_D + int(float64(BURST_SIZE_D)*dipVirionRatio)
								}
								//  ---------------------------------------
								// Partition mode: split particles between random jump and cell-to-cell

								randomVirions := int(math.Floor(float64(BURST_SIZE_V) * k_JumpR))
								virionsForLocalDiffusion := BURST_SIZE_V - randomVirions

								randomDIPs := int(math.Floor(float64(adjustedBurstSizeD) * k_JumpR))
								dipsForLocalDiffusion := adjustedBurstSizeD - randomDIPs

								// Handle random jumps
								for v := 0; v < randomVirions; v++ {
									ni, nj := rand.Intn(GRID_SIZE), rand.Intn(GRID_SIZE)
									g.localVirions[ni][nj]++
									g.totalRandomJumpVirions++
								}
								for d := 0; d < randomDIPs; d++ {
									ni, nj := rand.Intn(GRID_SIZE), rand.Intn(GRID_SIZE)
									g.localDips[ni][nj]++
									g.totalRandomJumpDIPs++
								}

								// Handle local diffusion
								// Handle local diffusion with localVirions & localDIPs (keep original logic unchanged)
								if virionsForLocalDiffusion > 0 || dipsForLocalDiffusion > 0 {
									// Calculate the total number of valid neighbors
									totalNeighbors := 0

									// Count valid neighbors from neighbors1
									for _, dir := range g.neighbors1[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}
									// Count valid neighbors from neighbors2
									for _, dir := range g.neighbors2[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}
									// Count valid neighbors from neighbors3
									for _, dir := range g.neighbors3[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}

									if totalNeighbors == 0 {
										return
									}

									// Calculate the distribution based on the ratio √3 : 2√3 : 3
									sqrt3 := math.Sqrt(3)
									ratio1 := 1.0               // sqrt3     // Weight for neighbors1
									ratio2 := 1.0 / 2           // 2 * sqrt3 // Weight for neighbors2
									ratio3 := 1.0 / (3 / sqrt3) // 3.0       // Weight for neighbors3
									totalRatio := ratio1*float64(len(g.neighbors1[i][j])) +
										ratio2*float64(len(g.neighbors2[i][j])) +
										ratio3*float64(len(g.neighbors3[i][j]))

									// Calculate virions for each neighbor group
									virionsForNeighbors1 := int(math.Floor(float64(virionsForLocalDiffusion) * (ratio1 * float64(len(g.neighbors1[i][j]))) / totalRatio))
									virionsForNeighbors2 := int(math.Floor(float64(virionsForLocalDiffusion) * (ratio2 * float64(len(g.neighbors2[i][j]))) / totalRatio))
									virionsForNeighbors3 := int(math.Floor(float64(virionsForLocalDiffusion) * (ratio3 * float64(len(g.neighbors3[i][j]))) / totalRatio))

									// Calculate remaining virions
									remainingVirions := virionsForLocalDiffusion - (virionsForNeighbors1 + virionsForNeighbors2 + virionsForNeighbors3)

									// Distribute remaining virions based on ratio
									for remainingVirions > 0 {
										randVal := rand.Float64() * totalRatio
										if randVal < ratio1 && len(g.neighbors1[i][j]) > 0 {
											virionsForNeighbors1++
										} else if randVal < (ratio1+ratio2) && len(g.neighbors2[i][j]) > 0 {
											virionsForNeighbors2++
										} else if len(g.neighbors3[i][j]) > 0 {
											virionsForNeighbors3++
										}
										remainingVirions--
									}

									// Calculate DIPs for each neighbor group (same logic)
									dipsForNeighbors1 := int(math.Floor(float64(dipsForLocalDiffusion) * (ratio1 * float64(len(g.neighbors1[i][j]))) / totalRatio))
									dipsForNeighbors2 := int(math.Floor(float64(dipsForLocalDiffusion) * (ratio2 * float64(len(g.neighbors2[i][j]))) / totalRatio))
									dipsForNeighbors3 := int(math.Floor(float64(dipsForLocalDiffusion) * (ratio3 * float64(len(g.neighbors3[i][j]))) / totalRatio))

									remainingDIPs := dipsForLocalDiffusion - (dipsForNeighbors1 + dipsForNeighbors2 + dipsForNeighbors3)

									for remainingDIPs > 0 {
										randVal := rand.Float64() * totalRatio
										if randVal < ratio1 && len(g.neighbors1[i][j]) > 0 {
											dipsForNeighbors1++
										} else if randVal < (ratio1+ratio2) && len(g.neighbors2[i][j]) > 0 {
											dipsForNeighbors2++
										} else if len(g.neighbors3[i][j]) > 0 {
											dipsForNeighbors3++
										}
										remainingDIPs--
									}
									// Distribute virions to neighbors1
									for _, dir := range g.neighbors1[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											if g.state[ni][nj] == SUSCEPTIBLE {
												g.localVirions[ni][nj] += virionsForNeighbors1 / len(g.neighbors1[i][j])
												g.localDips[ni][nj] += dipsForNeighbors1 / len(g.neighbors1[i][j])
											}
										}
									}

									// Distribute virions to neighbors2
									for _, dir := range g.neighbors2[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											g.localVirions[ni][nj] += virionsForNeighbors2 / len(g.neighbors2[i][j])
											g.localDips[ni][nj] += dipsForNeighbors2 / len(g.neighbors2[i][j])
										}
									}

									// Distribute virions to neighbors3
									for _, dir := range g.neighbors3[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											g.localVirions[ni][nj] += virionsForNeighbors3 / len(g.neighbors3[i][j])
											g.localDips[ni][nj] += dipsForNeighbors3 / len(g.neighbors3[i][j])
										}
									}
								}

							} else if par_celltocell_random == false {
								if !allowVirionJump && !allowDIPJump {
									// Calculate the total number of valid neighbors
									totalNeighbors := 0

									// Count valid neighbors from neighbors1
									for _, dir := range g.neighbors1[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}
									// Count valid neighbors from neighbors2
									for _, dir := range g.neighbors2[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}
									// Count valid neighbors from neighbors3
									for _, dir := range g.neighbors3[i][j] {
										if dir != [2]int{-1, -1} {
											totalNeighbors++
										}
									}

									// If there are no valid neighbors, return early
									if totalNeighbors == 0 {
										return
									}

									// Calculate the distribution of virions and DIPs to each neighbor based on the ratio √3 : 2√3 : 3
									sqrt3 := math.Sqrt(3)
									ratio1 := 1.0               // sqrt3     // Weight for neighbors1
									ratio2 := 1.0 / 2           // 2 * sqrt3 // Weight for neighbors2
									ratio3 := 1.0 / (3 / sqrt3) // 3.0 // Weight for neighbors3
									totalRatio := ratio1*float64(len(g.neighbors1[i][j])) + ratio2*float64(len(g.neighbors2[i][j])) + ratio3*float64(len(g.neighbors3[i][j]))

									// if infected by virion or infected by both:
									// Calculate the number of virions and DIPs assigned to each type of neighbor
									virionsForNeighbors1 := int(math.Floor(float64(BURST_SIZE_V) * (ratio1 * float64(len(g.neighbors1[i][j]))) / totalRatio))
									virionsForNeighbors2 := int(math.Floor(float64(BURST_SIZE_V) * (ratio2 * float64(len(g.neighbors2[i][j]))) / totalRatio))
									virionsForNeighbors3 := int(math.Floor(float64(BURST_SIZE_V) * (ratio3 * float64(len(g.neighbors3[i][j]))) / totalRatio))

									// Calculate the remaining virions and DIPs
									remainingVirions := BURST_SIZE_V - (virionsForNeighbors1 + virionsForNeighbors2 + virionsForNeighbors3)

									// // Randomly distribute the remaining virions based on the ratio
									for remainingVirions > 0 {
										randVal := rand.Float64() * totalRatio
										if randVal < ratio1 && len(g.neighbors1[i][j]) > 0 {
											virionsForNeighbors1++
										} else if randVal < (ratio1+ratio2) && len(g.neighbors2[i][j]) > 0 {
											virionsForNeighbors2++
										} else if len(g.neighbors3[i][j]) > 0 {
											virionsForNeighbors3++
										}
										remainingVirions--
									}
									// if infected by vrion only or both:

									totalVirionsAtCell := g.localVirions[i][j]
									totalDIPsAtCell := g.localDips[i][j]

									// Ensure we avoid division by zero
									adjustedBurstSizeD := 0
									if totalVirionsAtCell > 0 {
										// Adjust BURST_SIZE_D based on the DIP-to-virion ratio at this cell
										dipVirionRatio := float64(totalDIPsAtCell) / float64(totalVirionsAtCell)
										adjustedBurstSizeD = BURST_SIZE_D + int(math.Floor(float64(BURST_SIZE_D)*dipVirionRatio))
									}

									// Distribute DIPs to neighbors based on the adjusted BURST_SIZE_D
									dipsForNeighbors1 := int(math.Floor(float64(adjustedBurstSizeD) * (ratio1 * float64(len(g.neighbors1[i][j]))) / totalRatio))
									dipsForNeighbors2 := int(math.Floor(float64(adjustedBurstSizeD) * (ratio2 * float64(len(g.neighbors2[i][j]))) / totalRatio))
									dipsForNeighbors3 := int(math.Floor(float64(adjustedBurstSizeD) * (ratio3 * float64(len(g.neighbors3[i][j]))) / totalRatio))
									remainingDips := adjustedBurstSizeD - (dipsForNeighbors1 + dipsForNeighbors2 + dipsForNeighbors3)

									// Randomly distribute the remaining DIPs based on the ratio
									for remainingDips > 0 {
										randVal := rand.Float64() * totalRatio
										if randVal < ratio1 && len(g.neighbors1[i][j]) > 0 {
											dipsForNeighbors1++
										} else if randVal < (ratio1+ratio2) && len(g.neighbors2[i][j]) > 0 {
											dipsForNeighbors2++
										} else if len(g.neighbors3[i][j]) > 0 {
											dipsForNeighbors3++
										}
										remainingDips--
									}
									// Distribute virions and DIPs to neighbors1
									for _, dir := range g.neighbors1[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											if g.state[ni][nj] == SUSCEPTIBLE {
												g.localVirions[ni][nj] += virionsForNeighbors1 / len(g.neighbors1[i][j])
												g.localDips[ni][nj] += dipsForNeighbors1 / len(g.neighbors1[i][j])
											}
										}
									}

									// Distribute virions and DIPs to neighbors2
									for _, dir := range g.neighbors2[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											g.localVirions[ni][nj] += virionsForNeighbors2 / len(g.neighbors2[i][j])
											g.localDips[ni][nj] += dipsForNeighbors2 / len(g.neighbors2[i][j])
										}
									}

									// Distribute virions and DIPs to neighbors3
									for _, dir := range g.neighbors3[i][j] {
										ni, nj := dir[0], dir[1]
										if dir != [2]int{-1, -1} && ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
											g.localVirions[ni][nj] += virionsForNeighbors3 / len(g.neighbors3[i][j])
											g.localDips[ni][nj] += dipsForNeighbors3 / len(g.neighbors3[i][j])
										}
									}

								} else { // "Jump" case for either virions, DIPs, or both

									if allowVirionJump {
										totalVirionsAtCell := g.localVirions[i][j]
										totalDIPsAtCell := g.localDips[i][j]
										adjustedBurstSizeD := 0
										if totalVirionsAtCell > 0 {
											dipVirionRatio := float64(totalDIPsAtCell) / float64(totalVirionsAtCell)
											adjustedBurstSizeD = BURST_SIZE_D + int(math.Floor(float64(BURST_SIZE_D)*dipVirionRatio))
										}

										if jumpRandomly {
											for v := 0; v < BURST_SIZE_V; v++ {
												ni := rand.Intn(GRID_SIZE) // Randomly select a row
												nj := rand.Intn(GRID_SIZE) // Randomly select a column

												// Apply the virion jump
												g.localVirions[ni][nj]++
											}

											// DIP jump randomly to any location
											for d := 0; d < adjustedBurstSizeD; d++ {
												ni := rand.Intn(GRID_SIZE) // Randomly select a row
												nj := rand.Intn(GRID_SIZE) // Randomly select a column

												// Apply the DIP jump
												g.localDips[ni][nj]++
											}
										} else {
											// Virion jump logic
											virionTargets := make([]int, BURST_SIZE_V)
											for v := 0; v < BURST_SIZE_V; v++ {
												virionTargets[v] = rand.Intn(len(g.neighborsRingVirion[i][j]))
											}

											// Apply virion jumps
											for _, targetIndex := range virionTargets {
												spot := g.neighborsRingVirion[i][j][targetIndex]
												ni, nj := spot[0], spot[1]

												// Ensure the jump target is valid
												if ni < 0 || ni >= GRID_SIZE || nj < 0 || nj >= GRID_SIZE {
													// fmt.Printf("Skipping invalid jump target (%d, %d) from (%d, %d)\n", ni, nj, i, j)
													continue
												}

												// Apply the virion jump
												g.localVirions[ni][nj]++
											}

											// DIP jump logic
											dipTargets := make([]int, adjustedBurstSizeD)
											for d := 0; d < adjustedBurstSizeD; d++ {
												dipTargets[d] = rand.Intn(len(g.neighborsRingDIP[i][j]))
											}

											// Apply DIP jumps
											for _, targetIndex := range dipTargets {
												spot := g.neighborsRingDIP[i][j][targetIndex]
												ni, nj := spot[0], spot[1]

												// Ensure the jump target is valid
												if ni < 0 || ni >= GRID_SIZE || nj < 0 || nj >= GRID_SIZE {
													//fmt.Printf("Skipping invalid jump target (%d, %d) from (%d, %d)\n", ni, nj, i, j)
													continue
												}

												// Apply the DIP jump
												g.localDips[ni][nj]++
											}
										}
									}

									if allowDIPJump {
										totalVirionsAtCell := g.localVirions[i][j]
										totalDIPsAtCell := g.localDips[i][j]
										adjustedBurstSizeD := 0
										if totalVirionsAtCell > 0 {
											dipVirionRatio := float64(totalDIPsAtCell) / float64(totalVirionsAtCell)
											adjustedBurstSizeD = BURST_SIZE_D + int(math.Floor(float64(BURST_SIZE_D)*dipVirionRatio))
										}

										if jumpRandomly {
											go func() {
												for d := 0; d < adjustedBurstSizeD; d++ {
													ni := rand.Intn(GRID_SIZE)
													nj := rand.Intn(GRID_SIZE)
													g.localDips[ni][nj]++
												}
											}()
										} else {
											dipTargets := make([]int, adjustedBurstSizeD)
											for d := 0; d < adjustedBurstSizeD; d++ {
												dipTargets[d] = rand.Intn(len(g.neighborsRingDIP[i][j]))
											}
											go func() {
												for _, targetIndex := range dipTargets {
													spot := g.neighborsRingDIP[i][j][targetIndex]
													ni, nj := spot[0], spot[1]
													if ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {
														g.localDips[ni][nj]++
													}
												}
											}()
										}
									}
								}
							}
						}
					}
					// update infected only by DIP or only by virions cells become infected by both
					if g.state[i][j] == INFECTED_VIRION || g.state[i][j] == INFECTED_DIP {

						if g.stateChanged[i][j] == false {
							// Check if the cell is infected by virions or DIPs

							if g.localVirions[i][j] > 0 || g.localDips[i][j] > 0 {
								// Calculate the infection probabilities

								if R == 0 || TAU == 0 {
									perParticleInfectionChance_V = RHO
								} else {
									if VStimulateIFN == true { // R=1
										perParticleInfectionChance_V = RHO * math.Exp(-ALPHA*(globalIFNperCell/float64(R)))
									} else if VStimulateIFN == false { // usually only DIP stimulate IFN in this situlation
										perParticleInfectionChance_V = RHO * math.Exp(-ALPHA*(globalIFNperCell))
									}
								}
								var probabilityVInfection, probabilityDInfection float64

								// Virion infection probability
								probabilityVInfection = 1 - math.Pow(1-perParticleInfectionChance_V, float64(g.localVirions[i][j]))
								infectedByVirion := rand.Float64() <= probabilityVInfection

								// DIP infection probability
								probabilityDInfection = 1 - math.Pow(1-(RHO*math.Exp(-ALPHA*(globalIFNperCell))), float64(g.localDips[i][j]))
								infectedByDip := rand.Float64() <= probabilityDInfection

								// Determine the infection state based on virion and DIP infection
								if infectedByVirion && infectedByDip {
									newGrid[i][j] = INFECTED_BOTH
								} else if infectedByVirion {
									newGrid[i][j] = INFECTED_VIRION
								} else if infectedByDip {
									newGrid[i][j] = INFECTED_DIP
								}
							}

						}
						if g.state[i][j] == INFECTED_VIRION || g.state[i][j] == INFECTED_BOTH && TAU > 0 {

							if VStimulateIFN == true {
								if g.state[i][j] == INFECTED_VIRION {
									g.IFNConcentration[i][j] += float64(R) * float64(TIMESTEP) * ifnBothFold
								} else if g.state[i][j] == INFECTED_BOTH {

									// if g.intraWT[i][j] > 0 {
									// 	dvgWtRatio := float64(g.intraDVG[i][j]) / float64(g.intraWT[i][j])
									// 	if g.state[i][j] == INFECTED_BOTH {
									// 		adjusted_DIP_IFN_stimulate *= dvgWtRatio * BOTH_IFN_stimulate_ratio
									// 	}
									// }
									adjusted_DIP_IFN_stimulate = BOTH_IFN_stimulate_ratio
									g.IFNConcentration[i][j] += (float64(R) + adjusted_DIP_IFN_stimulate) * float64(TIMESTEP)
								}
							} else if VStimulateIFN == false {
								if g.state[i][j] == INFECTED_VIRION {
									// do nothing since virions do not stimulate IFN
								} else if g.state[i][j] == INFECTED_BOTH {

									// if g.intraWT[i][j] > 0 {
									// 	dvgWtRatio := float64(g.intraDVG[i][j]) / float64(g.intraWT[i][j])
									// 	if g.state[i][j] == INFECTED_BOTH {
									// 		adjusted_DIP_IFN_stimulate *= dvgWtRatio * BOTH_IFN_stimulate_ratio
									// 	}
									// }

									adjusted_DIP_IFN_stimulate = BOTH_IFN_stimulate_ratio

								}
								g.IFNConcentration[i][j] += (float64(R) + adjusted_DIP_IFN_stimulate) * float64(TIMESTEP)
							}

							globalIFN += g.IFNConcentration[i][j]

						}

						if g.state[i][j] == INFECTED_DIP {

							g.timeSinceInfectDIP[i][j] += TIMESTEP

							if g.timeSinceInfectDIP[i][j] > IFN_DELAY+int(math.Floor(rand.NormFloat64()*float64(STD_IFN_DELAY))) && TAU > 0 {

								//adjusted_DIP_IFN_stimulate := float64(g.intraDVG[i][j]) * D_only_IFN_stimulate_ratio
								adjusted_DIP_IFN_stimulate := D_only_IFN_stimulate_ratio
								g.IFNConcentration[i][j] += (float64(R) + adjusted_DIP_IFN_stimulate) * float64(TIMESTEP)
								globalIFN += g.IFNConcentration[i][j]
							}

						}

					}

				}
			}
		}
		// Handle potentially regrowing dead cells
		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				if g.state[i][j] == DEAD {
					g.timeSinceDead[i][j] += TIMESTEP

					// Check if any neighboring cells are susceptible, allowing for regrowth
					canRegrow := false
					neighbors := g.neighbors1[i][j]

					// Iterate over the neighbors and check if any are SUSCEPTIBLE
					for _, neighbor := range neighbors {
						ni, nj := neighbor[0], neighbor[1]

						// Ensure the neighbor indices are valid (within grid bounds)
						if ni >= 0 && ni < GRID_SIZE && nj >= 0 && nj < GRID_SIZE {

							if g.state[ni][nj] == SUSCEPTIBLE || g.state[ni][nj] == ANTIVIRAL {
								canRegrow = true
								break

							}

						}
					}

					// If the conditions are met, the cell regrows
					if canRegrow && g.timeSinceDead[i][j] >= int(rand.NormFloat64()*REGROWTH_STD+REGROWTH_MEAN) {
						newGrid[i][j] = REGROWTH
						g.timeSinceRegrowth[i][j] = 0
						g.timeSinceDead[i][j] = -1

					}

				}
			}
		}
		// IFN exponential decay

		if ifn_half_life != 0 {
			globalIFN = globalIFN * math.Pow(0.5, float64(TIMESTEP)/ifn_half_life)
			if globalIFN < (1.0 / (float64(GRID_SIZE) * float64(GRID_SIZE))) {
				globalIFN = 0
			}
		}

		globalIFNperCell = globalIFN / float64(GRID_SIZE*GRID_SIZE)
		// Apply the updated grid state
		g.state = newGrid

		// Calculate and log the total virions and DIPs for each time step
		totalVirions, totalDIPs := g.totalVirions(), g.totalDIPs()
		fmt.Printf("Time step %d: Total Virions = %d, Total DIPs = %d\n", frameNum, totalVirions, totalDIPs)

		// Additional calculations based on simulation parameters for tracking purposes
		regrowthCount := g.calculateRegrowthCount()
		susceptiblePercentage := g.calculateSusceptiblePercentage()

		regrowthedOrAntiviralPercentage := g.calculateRegrowthedOrAntiviralPercentage()
		infectedPercentage := g.calculateInfectedPercentage()
		infectedDIPOnlyPercentage := g.calculateInfectedDIPOnlyPercentage()
		infectedBothPercentage := g.calculateInfectedBothPercentage()
		antiviralPercentage := g.calculateAntiviralPercentage()
		deadCellPercentage := calculateDeadCellPercentage(g.state)
		uninfectedPercentage := g.calculateUninfectedPercentage()
		plaquePercentage := g.calculatePlaquePercentage()
		//virionDiffusionRate, dipDiffusionRate := g.calculateDiffusionRates()

		// Log additional data as necessary
		fmt.Printf("Regrowth Count: %d, Susceptible: %.2f%%", regrowthCount, susceptiblePercentage)
		fmt.Printf("Regrowthed or Antiviral: %.2f%%, Infected: %.2f%%, DIP Only: %.2f%%, Both Infected: %.2f%%, Antiviral: %.2f%%\n",
			regrowthedOrAntiviralPercentage, infectedPercentage, infectedDIPOnlyPercentage, infectedBothPercentage, antiviralPercentage)
		fmt.Printf("Dead: %.2f%%, Uninfected: %.2f%%, Plaque: %.2f%%\n", deadCellPercentage, uninfectedPercentage, plaquePercentage)
		//fmt.Printf("Virion Diffusion Rate: %d, DIP Diffusion Rate: %d\n", virionDiffusionRate, dipDiffusionRate)

	}

	// TIMESTEP = 1 hour. If 1 hour/step, use dt = 1.0

	if virion_half_life != 0 {
		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				// Update virus count using half-life formula
				factorV := math.Pow(0.5, float64(TIMESTEP)/virion_half_life)
				g.localVirions[i][j] = int(math.Floor(float64(g.localVirions[i][j])*factorV + 0.5))

				if dip_half_life != 0 {
					factorD := math.Pow(0.5, float64(TIMESTEP)/dip_half_life)
					g.localDips[i][j] = int(math.Floor(float64(g.localDips[i][j])*factorD + 0.5))
				}
			}
		}
	}

}

// Function to record simulation data into CSV at each timestep
// Function to record simulation data into CSV at each timestep
func (g *Grid) recordSimulationData(writer *csv.Writer, frameNum int) {
	totalVirions := g.totalVirions()
	totalDIPs := g.totalDIPs()
	deadCellPercentage := strconv.FormatFloat(calculateDeadCellPercentage(g.state), 'f', 6, 64)
	susceptiblePercentage := strconv.FormatFloat(g.calculateSusceptiblePercentage(), 'f', 6, 64)
	infectedPercentage := strconv.FormatFloat(g.calculateInfectedPercentage(), 'f', 6, 64)
	infectedDIPOnlyPercentage := strconv.FormatFloat(g.calculateInfectedDIPOnlyPercentage(), 'f', 6, 64)
	infectedBothPercentage := strconv.FormatFloat(g.calculateInfectedBothPercentage(), 'f', 6, 64)
	antiviralPercentage := strconv.FormatFloat(g.calculateAntiviralPercentage(), 'f', 6, 64)
	virionOnlyInfected := g.calculateVirionOnlyInfected()
	dipOnlyInfected := g.calculateDipOnlyInfected()
	bothInfected := g.calculateBothInfected()

	// Calculate DIP advantage = burstSizeD / burstSizeV
	dipAdvantage = float64(BURST_SIZE_D) / float64(BURST_SIZE_V)

	row := []string{
		strconv.Itoa(frameNum),
		strconv.FormatFloat(virion_half_life, 'f', 6, 64), // Add virion clearance rate
		strconv.FormatFloat(dip_half_life, 'f', 6, 64),    // Add DIP clearance rate
		strconv.FormatFloat(ifn_half_life, 'f', 6, 64),    // Add IFN clearance rate
		strconv.FormatFloat(globalIFN/float64(GRID_SIZE*GRID_SIZE), 'f', 6, 64),
		strconv.Itoa(totalVirions),
		strconv.Itoa(totalDIPs),
		deadCellPercentage,
		susceptiblePercentage,
		infectedPercentage,
		infectedDIPOnlyPercentage,
		infectedBothPercentage,
		antiviralPercentage,
		strconv.Itoa(g.calculateRegrowthCount()),
		strconv.FormatFloat(g.calculateSusceptiblePercentage(), 'f', 6, 64),
		strconv.FormatFloat(g.calculateRegrowthedOrAntiviralPercentage(), 'f', 6, 64),
		"variate, depending on radius 10 of IFN",
		"variate, depending on radius 10 of IFN",
		strconv.FormatFloat(RHO, 'f', 6, 64),
		strconv.Itoa(totalVirions + totalDIPs),
		strconv.FormatFloat(g.calculatePlaquePercentage(), 'f', 6, 64),
		strconv.FormatFloat(float64(maxGlobalIFN), 'f', 6, 64),
		"-1.0",
		strconv.FormatFloat(g.calculateUninfectedPercentage(), 'f', 6, 64),
		"0",
		strconv.Itoa(GRID_SIZE),
		strconv.Itoa(TIMESTEP),
		strconv.Itoa(IFN_DELAY),
		strconv.Itoa(STD_IFN_DELAY),
		strconv.FormatFloat(ALPHA, 'f', 6, 64),
		strconv.FormatFloat(RHO, 'f', 6, 64),
		strconv.FormatFloat(float64(TAU), 'f', 6, 64),
		strconv.Itoa(BURST_SIZE_V),
		strconv.FormatFloat(REGROWTH_MEAN, 'f', 6, 64),
		strconv.FormatFloat(REGROWTH_STD, 'f', 6, 64),
		strconv.Itoa(TIME_STEPS),
		strconv.FormatFloat(MEAN_LYSIS_TIME, 'f', 6, 64),
		strconv.FormatFloat(STANDARD_LYSIS_TIME, 'f', 6, 64),
		strconv.FormatFloat(float64(*flag_v_pfu_initial)/float64(GRID_SIZE*GRID_SIZE), 'f', 6, 64),
		strconv.FormatFloat(float64(*flag_d_pfu_initial)/float64(GRID_SIZE*GRID_SIZE), 'f', 6, 64),
		"-1.0",
		"-1.0",
		strconv.FormatFloat(float64(R), 'f', 6, 64),
		strconv.Itoa(BURST_SIZE_D),
		"-1.0",

		strconv.Itoa(option),
		strconv.FormatFloat(*flag_v_pfu_initial, 'f', -1, 64),
		strconv.FormatFloat(*flag_d_pfu_initial, 'f', -1, 64),
		strconv.Itoa(virionOnlyInfected),
		strconv.Itoa(dipOnlyInfected),
		strconv.Itoa(bothInfected),
		strconv.Itoa(totalDeadFromV),
		strconv.Itoa(totalDeadFromBoth),
		strconv.Itoa(virionDiffusionRate),
		strconv.Itoa(dipDiffusionRate),
		strconv.FormatFloat(k_JumpR, 'f', 6, 64),
		strconv.Itoa(jumpRadiusV),
		strconv.Itoa(jumpRadiusD),
		strconv.FormatBool(jumpRandomly),
		strconv.FormatBool(par_celltocell_random),
		strconv.FormatBool(allowVirionJump),
		strconv.FormatBool(allowDIPJump),
		strconv.Itoa(IFN_wave_radius),
		strconv.FormatBool(ifnWave),
		strconv.FormatFloat(ifnBothFold, 'f', 6, 64),
		strconv.FormatFloat(D_only_IFN_stimulate_ratio, 'f', 6, 64),
		strconv.FormatFloat(BOTH_IFN_stimulate_ratio, 'f', 6, 64),
		strconv.Itoa(g.totalRandomJumpVirions),        // New: total number of randomly jumping Virions
		strconv.Itoa(g.totalRandomJumpDIPs),           // New: total number of randomly jumping DIPs
		strconv.FormatFloat(dipAdvantage, 'f', 6, 64), // DIP advantage = burstSizeD / burstSizeV
	}

	writer.Write(row)
	writer.Flush()
}

// Convert the grid state into an image
func (g *Grid) gridToImage(videotype string) *image.RGBA {

	imgWidth := GRID_SIZE * CELL_SIZE * 2                       // Calculate the image width
	imgHeight := GRID_SIZE * CELL_SIZE * 2                      // Calculate the image height
	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight)) // Create a new image
	if videotype == "states" {
		// Define colors for different states
		colors := map[int]color.Color{
			SUSCEPTIBLE:     color.RGBA{0, 0, 0, 255},       // Susceptible state: black
			INFECTED_VIRION: color.RGBA{255, 0, 0, 255},     // Infected by virion: red
			INFECTED_DIP:    color.RGBA{0, 255, 0, 255},     // Infected by DIP: green
			INFECTED_BOTH:   color.RGBA{255, 255, 0, 255},   // Infected by both: yellow
			DEAD:            color.RGBA{169, 169, 169, 255}, // Dead state: gray
			ANTIVIRAL:       color.RGBA{0, 0, 255, 255},     // Antiviral state: blue
			REGROWTH:        color.RGBA{128, 0, 128, 255},   // Regrowth state: purple
		}
		fillBackground(img, color.RGBA{0, 0, 0, 255})
		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				x, y := calculateHexCenter(i, j)              // Calculate the center of each hexagon
				drawHexagon(img, x, y, colors[g.state[i][j]]) // Draw the hexagon based on the cell state
			}
		}
		// Return the image
	} else if videotype == "IFNconcentration" { // IFN concentration visualization
		black := color.RGBA{0, 0, 0, 255} // Default color (black)

		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				x, y := calculateHexCenter(i, j) // Calculate hexagon center coordinates
				ifnValue := g.IFNConcentration[i][j]

				var cellColor color.RGBA
				if ifnValue <= 0 {
					cellColor = black // IFN ≤ 0, black
				} else if ifnValue > 0 && ifnValue <= 1 {
					cellColor = color.RGBA{0, 0, 255, 255} // Blue
				} else if ifnValue > 1 && ifnValue <= 2 {
					cellColor = color.RGBA{0, 255, 0, 255} // Green
				} else if ifnValue > 2 && ifnValue <= 5 {
					cellColor = color.RGBA{255, 255, 0, 255} // Yellow
				} else if ifnValue > 5 && ifnValue <= 10 {
					cellColor = color.RGBA{255, 165, 0, 255} // Orange
				} else {
					cellColor = color.RGBA{255, 0, 0, 255} // Red
				}

				drawHexagon(img, x, y, cellColor)
			}
		}
	} else if videotype == "IFNonlyLargerThanZero" { // IFN concentration visualization
		red := color.RGBA{255, 0, 0, 255} // Cells with interferon > 0
		blue := color.RGBA{0, 0, 255, 255}
		black := color.RGBA{0, 0, 0, 255} // Default color for all other cells
		yellow := color.RGBA{255, 255, 0, 255}
		green := color.RGBA{0, 255, 0, 255}
		organge := color.RGBA{255, 165, 0, 255}
		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				x, y := calculateHexCenter(i, j) // Calculate the center of each hexagon

				// Apply color based on the specified conditions
				if g.timeSinceAntiviral[i][j] > g.antiviralDuration[i][j] {
					drawHexagon(img, x, y, blue) // blue for cells in antiviral state exceeding duration

				} else if g.timeSinceAntiviral[i][j] > 110 {
					drawHexagon(img, x, y, red) //

				} else if g.timeSinceAntiviral[i][j] > 90 {
					drawHexagon(img, x, y, organge) //

				} else if g.timeSinceAntiviral[i][j] > 70 {
					drawHexagon(img, x, y, green) //

				} else if g.timeSinceAntiviral[i][j] > 50 {
					drawHexagon(img, x, y, yellow) //

				} else {
					drawHexagon(img, x, y, black) // Black for all other cells
				}
			}
		}
	} else if videotype == "antiviralState" {

		blue := color.RGBA{0, 0, 255, 255}
		black := color.RGBA{0, 0, 0, 255} // Default color for all other cells

		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				x, y := calculateHexCenter(i, j) // Calculate the center of each hexagon

				// Apply color based on the specified conditions
				if g.timeSinceAntiviral[i][j] > g.antiviralDuration[i][j] {
					drawHexagon(img, x, y, blue) // blue for cells in antiviral state exceeding duration
				} else {
					drawHexagon(img, x, y, black) // Black for all other cells
				}
			}
		}
	} else if videotype == "particles" {

		fillBackground(img, color.RGBA{0, 0, 0, 255})
		for i := 0; i < GRID_SIZE; i++ {
			for j := 0; j < GRID_SIZE; j++ {
				x, y := calculateHexCenter(i, j)

				// Determine color based on particle presence
				hasVirion := g.localVirions[i][j] > 0
				hasDIP := g.localDips[i][j] > 0

				var particleColor color.Color
				switch {
				case hasVirion && hasDIP:
					particleColor = color.RGBA{255, 255, 0, 255} // Yellow (both present)
				case hasVirion:
					particleColor = color.RGBA{255, 0, 0, 255} // Red (Virion only)
				case hasDIP:
					particleColor = color.RGBA{0, 255, 0, 255} // Green (DIP only)
				default:
					particleColor = color.RGBA{0, 0, 0, 255} // Black (no particles)
				}

				drawHexagon(img, x, y, particleColor)

				// Optional: add particle count text at hexagon center
				//if hasVirion || hasDIP {
				//	label := fmt.Sprintf("V:%d\nD:%d", g.localVirions[i][j], g.localDips[i][j])
				//	addLabelCentered(img, x, y, label, color.White)
				//}

			}
		}

	} else {
		fmt.Println("Error: Unknown videotype provided.")
	}

	return img // Return the image
}

func drawTextWithBackground(img *image.RGBA, x, y int, label string, textColor, borderColor, bgColor color.Color) {
	face := basicfont.Face7x13
	textWidth := len(label) * 7
	textHeight := 13

	// White background box
	bgRect := image.Rect(x-4, y-4, x+textWidth+4, y+textHeight+4)
	draw.Draw(img, bgRect, &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Text starting point
	point := fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y + textHeight),
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: face,
		Dot:  point,
	}
	d.DrawString(label)
}

// addLabel draws a text label onto an image at the specified position.
func addLabel(img *image.RGBA, x, y int, label string, col color.Color) {
	point := fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13, // Basic font for rendering
		Dot:  point,
	}
	d.DrawString(label)
}
func addStaticLegend(img *image.RGBA, startX, startY int) {
	// Keep original colors and label definitions unchanged
	legendItems := []string{
		"By both", "By DIP", "By Virion",
		"Antiviral", "Uninfected", "Plaque", "Regrowth",
	}
	legendColors := map[string]color.Color{
		"By both":    color.RGBA{255, 200, 0, 255},
		"By DIP":     color.RGBA{0, 255, 0, 255},
		"By Virion":  color.RGBA{255, 0, 0, 255},
		"Antiviral":  color.RGBA{0, 102, 255, 255},
		"Uninfected": color.RGBA{0, 0, 0, 255},
		"Plaque":     color.RGBA{84, 110, 122, 255},
		"Regrowth":   color.RGBA{128, 0, 128, 255},
	}

	// Calculate background box size (keep original logic)
	const (
		fontWidth   = 7
		lineSpacing = 17
		padding     = 4
	)

	maxLabelLen := 0
	for _, label := range legendItems {
		if len(label) > maxLabelLen {
			maxLabelLen = len(label)
		}
	}

	bgWidth := maxLabelLen*fontWidth + 10
	bgHeight := len(legendItems)*lineSpacing + 6

	// Draw background (keep white opaque)
	bgRect := image.Rect(
		startX-padding,
		startY-padding,
		startX+bgWidth,
		startY+bgHeight,
	)
	draw.Draw(img, bgRect, &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.Point{}, draw.Src)

	// Draw legend items (keep original drawing logic)
	for i, label := range legendItems {
		yPos := startY + i*lineSpacing
		drawTextWithBackground(
			img,
			startX,
			yPos,
			label,
			legendColors[label],
			legendColors[label],
			color.RGBA{255, 255, 255, 255},
		)
	}
}

func (g *Grid) gridToImageWithGraph(frameNum int, virionOnly, dipOnly, both []float64, mode string, showLegend bool) *image.RGBA {
	const graphHeight = 100
	const spacing = 0

	gridImg := g.gridToImage(videotype)
	gridHeight := gridImg.Bounds().Dy()

	imgWidth := GRID_SIZE * CELL_SIZE * 2
	imgHeight := graphHeight + gridHeight + spacing
	canvas := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	graphImg := createInfectionGraph(frameNum, virionOnly, dipOnly, both, showLegend)
	draw.Draw(canvas, image.Rect(0, 0, imgWidth, graphHeight), graphImg, image.Point{}, draw.Src)
	draw.Draw(canvas, image.Rect(0, graphHeight+spacing, imgWidth, graphHeight+gridHeight+spacing), gridImg, image.Point{}, draw.Src)

	if showLegend {
		addStaticLegend(canvas, canvas.Bounds().Dx()-183, canvas.Bounds().Dy()-183)
	}

	return canvas
}

// Calculate the center of each hexagonal cell
func calculateHexCenter(i, j int) (int, int) {
	x := i * CELL_SIZE * 3 / 2                                                          // Calculate the x-coordinate
	y := int(float64(j)*CELL_SIZE*math.Sqrt(3) + float64(i%2)*CELL_SIZE*math.Sqrt(3)/2) // Calculate the y-coordinate
	return x, y                                                                         // Return the center coordinates
}

func drawHexagon(img *image.RGBA, x, y int, c color.Color) {
	var hex [6]image.Point
	for i := 0; i < 6; i++ {
		angle := math.Pi / 3 * float64(i) // Calculate the angle for each vertex of the hexagon
		hex[i] = image.Point{
			X: x + int(float64(CELL_SIZE)*math.Cos(angle)), // Calculate x-coordinate
			Y: y + int(float64(CELL_SIZE)*math.Sin(angle)), // Calculate y-coordinate
		}
	}
	fillHexagon(img, hex, c) // Fill the hexagon with the specified color
}

func fillHexagon(img *image.RGBA, hex [6]image.Point, c color.Color) {
	minX, minY, maxX, maxY := hex[0].X, hex[0].Y, hex[0].X, hex[0].Y // Initialize boundary values
	for _, p := range hex {
		if p.X < minX {
			minX = p.X // Update minimum x-coordinate
		}
		if p.Y < minY {
			minY = p.Y // Update minimum y-coordinate
		}
		if p.X > maxX {
			maxX = p.X // Update maximum x-coordinate
		}
		if p.Y > maxY {
			maxY = p.Y // Update maximum y-coordinate
		}
	}
	for x := minX; x <= maxX; x++ { // Iterate through x-coordinates
		for y := minY; y <= maxY; y++ { // Iterate through y-coordinates
			if isPointInHexagon(image.Point{x, y}, hex) { // Check if the point is inside the hexagon
				img.Set(x, y, c) // Set the color of the point
			}
		}
	}
}

func isPointInHexagon(p image.Point, hex [6]image.Point) bool {
	for i := 0; i < 6; i++ {
		j := (i + 1) % 6
		if (hex[j].X-hex[i].X)*(p.Y-hex[i].Y)-(hex[j].Y-hex[i].Y)*(p.X-hex[i].X) < 0 {
			return false // Return false if the point is outside the hexagon
		}
	}
	return true // Return true if the point is inside the hexagon
}

func main() {
	flag.Parse()
	fmt.Printf("Parsed ifnSpreadOption: %q\n", *flag_ifnSpreadOption)
	fmt.Printf("Parsed particleSpreadOption: %q\n", *flag_particleSpreadOption)

	// Assign parsed flag values to global variables (note dereferencing)
	BURST_SIZE_V = *flag_burstSizeV
	BURST_SIZE_D = *flag_burstSizeD
	MEAN_LYSIS_TIME = *flag_meanLysisTime
	STANDARD_LYSIS_TIME = MEAN_LYSIS_TIME / 4
	k_JumpR = *flag_kJumpR
	TAU = *flag_tau
	ifnBothFold = *flag_ifnBothFold
	RHO = *flag_rho
	option = *flag_option

	virion_half_life = *flag_virion_half_life
	dip_half_life = *flag_dip_half_life
	ifn_half_life = *flag_ifn_half_life

	particleSpreadOption = *flag_particleSpreadOption
	ifnSpreadOption = *flag_ifnSpreadOption
	dipOption = *flag_dipOption
	// Recalculate dependent parameters (note that ifnBothFold is now float64, not *float64)
	D_only_IFN_stimulate_ratio = 5.0 * ifnBothFold
	BOTH_IFN_stimulate_ratio = 10.0 * ifnBothFold
	videotype = *flag_videotype
	fmt.Printf("flag_videotype = %q\n", *flag_videotype)
	// Optional: print debug information
	fmt.Printf("Parameters:\n  burstSizeV = %d\n  burstSizeD = %d\n  MEAN_LYSIS_TIME = %.2f\n  kJumpR = %.2f\n  TAU = %d\n  ifnBothFold = %.2f\n  RHO = %.3f\n par_celltocell_random = %v\n",
		BURST_SIZE_V, BURST_SIZE_D, MEAN_LYSIS_TIME, k_JumpR, TAU, ifnBothFold, RHO, par_celltocell_random)

	// --- Particle Diffusion Options ---
	particleSpreadOption = *flag_particleSpreadOption
	if particleSpreadOption == "celltocell" {
		jumpRadiusV = 0
		jumpRadiusD = 0
		jumpRandomly = false
		// k_JumpR = 0.0
		allowVirionJump = false
		allowDIPJump = false
		fmt.Println("flag main celltocell")
	} else if particleSpreadOption == "jumprandomly" {
		jumpRadiusV = 0
		jumpRadiusD = 0
		jumpRandomly = true
		// par_celltocell_random = false
		allowVirionJump = true
		allowDIPJump = true
		// k_JumpR = 1.0
		fmt.Println("flag main jump randomly")
	} else if particleSpreadOption == "jumpradius" {
		jumpRadiusV = 5
		jumpRadiusD = 5
		jumpRandomly = false
		allowVirionJump = true
		allowDIPJump = true
		// k_JumpR = 0.0
	} else if particleSpreadOption == "partition" {
		jumpRadiusV = 0
		jumpRadiusD = 0
		jumpRandomly = true
		par_celltocell_random = true
		allowVirionJump = true // Need to enable jumping
		allowDIPJump = true    // Need to enable jumping
		fmt.Println("DEBUG: par_celltocell_random set to", par_celltocell_random)

		k_JumpR = *flag_kJumpR
	} else {
		log.Fatalf("Unknown particleSpreadOption: %s", particleSpreadOption)
	}
	fmt.Println("\nParticle spread option settings:")
	fmt.Printf("  particleSpreadOption: %s\n", particleSpreadOption)
	fmt.Printf("  jumpRadiusV: %d, jumpRadiusD: %d, jumpRandomly: %v, k_JumpR: %.2f\n",
		jumpRadiusV, jumpRadiusD, jumpRandomly, k_JumpR)

	// --- IFN Propagation Options ---
	ifnSpreadOption = *flag_ifnSpreadOption

	switch ifnSpreadOption {

	case "global":
		IFN_wave_radius = 0
		ifnWave = false
		fmt.Printf("hello: ifnSpreadOption set to: %s, IFN_wave_radius: %d\n", ifnSpreadOption, IFN_wave_radius)

	case "local":
		IFN_wave_radius = 10
		ifnWave = true
		fmt.Printf("ummmm: ifnSpreadOption set to: %s, IFN_wave_radius: %d\n", ifnSpreadOption, IFN_wave_radius)

	case "noIFN":
		IFN_wave_radius = 0
		// Disable IFN: set IFN-related parameters to zero
		ifnBothFold = 0.0
		// Additionally in the model, R, ALPHA, IFN_DELAY, STD_IFN_DELAY, tau, etc. can be set to zero
		ifnWave = false
		ALPHA = 0.0
		IFN_DELAY = 0
		STD_IFN_DELAY = 0
		TAU = 0
		ifn_half_life = 0.0
	default:
		log.Fatalf("Unknown ifnSpreadOption: %s", ifnSpreadOption)
		fmt.Printf("ifnSpreadOption set to: %s, IFN_wave_radius: %d\n", ifnSpreadOption, IFN_wave_radius)

	}
	fmt.Println("\nIFN spread option settings:")
	fmt.Printf("  ifnSpreadOption: %s, IFN_wave_radius: %d, ifnBothFold: %.2f\n",
		ifnSpreadOption, IFN_wave_radius, ifnBothFold)
	fmt.Printf("flag_ifnSpreadOption = %q\n", *flag_ifnSpreadOption)
	// --- DIP Options ---
	dipOption = *flag_dipOption
	if dipOption {
		BURST_SIZE_D = *flag_burstSizeD
		// Keep D_only_IFN_stimulate_ratio default value
	} else {
		BURST_SIZE_D = 0
		D_only_IFN_stimulate_ratio = 0.0
	}
	fmt.Println("\nDIP option settings:")
	fmt.Printf("  dipOption: %v, BURST_SIZE_D: %d, D_only_IFN_stimulate_ratio: %.2f, BOTH_IFN_stimulate_ratio: %.2f\n",
		dipOption, BURST_SIZE_D, D_only_IFN_stimulate_ratio, BOTH_IFN_stimulate_ratio)

	// Simulation code can be integrated here later, this example only shows parameter setup
	fmt.Println("\nSimulation initialization complete.")
	var grid Grid

	// Use current time as seed for reproducibility
	rand.Seed(time.Now().UnixNano())
	// Dynamically set the value of R
	if VStimulateIFN {
		R = int(1 * ifnBothFold)
	} else {
		R = 0
	}
	grid.initialize()                // Initialize the grid
	grid.initializeNeighbors()       // Initialize the neighbors
	grid.initializeInfection(option) // Initialize the infection state

	switch {
	case TIME_STEPS > 1000:
		ticksInterval = 500.0
	case TIME_STEPS == 145:
		ticksInterval = 24.0
	case TIME_STEPS > 500:
		ticksInterval = 100.0
	case TIME_STEPS > 100:
		ticksInterval = 50.0
	case TIME_STEPS%24 == 0:
		ticksInterval = 24.0
	default:
		ticksInterval = 100.0
	}

	// Switch statement with conditional cases
	// Switch statement with conditional cases
	switch {
	case IFN_wave_radius == 0 && TAU == 12 && jumpRandomly == true:
		yMax = 0.2
	case IFN_wave_radius == 0 && TAU == 12 && jumpRandomly == true:
		yMax = 1.0
	case IFN_wave_radius == 0 && TAU == 12 && jumpRandomly == true:
		yMax = 0.03
	case IFN_wave_radius == 0 && TAU == 12 && jumpRandomly == true:
		yMax = 1.5
	case IFN_wave_radius == 10 && TAU == 12 && jumpRandomly == true:
		yMax = 20.0
	case IFN_wave_radius == 10 && TAU == 12 && jumpRandomly == true:
		yMax = 0.1
	case IFN_wave_radius == 0 && jumpRadiusV == 0 && jumpRadiusD == 0 && TAU == 12:
		yMax = 0.3
	case IFN_wave_radius == 0 && jumpRadiusV == 5 && jumpRadiusD == 0 && TAU == 12:
		yMax = 1.0
	case IFN_wave_radius == 0 && jumpRadiusV == 0 && jumpRadiusD == 5 && TAU == 12:
		yMax = 0.03
	case IFN_wave_radius == 0 && jumpRadiusV == 5 && jumpRadiusD == 5 && TAU == 12:
		yMax = 0.1
	case IFN_wave_radius == 10 && jumpRadiusV == 5 && jumpRadiusD == 5 && TAU == 12:
		yMax = 0.2
	case IFN_wave_radius == 0 && jumpRadiusV == 0 && jumpRadiusD == 0 && TAU == 24:
		yMax = 0.3
	case IFN_wave_radius == 0 && jumpRadiusV == 5 && jumpRadiusD == 5 && TAU == 24:
		yMax = 1.5
	case IFN_wave_radius == 10 && jumpRadiusV == 5 && jumpRadiusD == 5 && TAU == 24:
		yMax = 0.2

	case IFN_wave_radius == 0 && jumpRadiusV == 0 && jumpRadiusD == 0:
		yMax = 0.3
	case IFN_wave_radius == 0 && jumpRadiusV == 5 && jumpRadiusD == 5:
		yMax = 1.5
	case IFN_wave_radius == 10 && jumpRadiusV == 5 && jumpRadiusD == 5:
		yMax = 35.0

	default:
		yMax = -1.0 // Default value in case no conditions are met
	}

	folderNumber := getNextFolderNumber("./")

	// Call generateFolderName function to generate folder name
	outputFolder := generateFolderName(
		folderNumber, // Current folder number
		jumpRandomly, // DIP random jumping logic
		jumpRadiusD,  // DIP jump radius
		jumpRadiusV,  // Virion jump radius
		BURST_SIZE_D, // DIP burst size
		BURST_SIZE_V, // Virion burst size
		//V_PFU_INITIAL,   // Virion initial value
		//D_PFU_INITIAL,   // DIP initial value
		IFN_wave_radius, // IFN wave radius
		TAU,             // TAU value
		TIME_STEPS,      // Time steps
	)

	// Create folder
	os.Mkdir(outputFolder, os.ModePerm)

	err := os.MkdirAll(outputFolder, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create folder: %v", err)
	}
	saveCurrentGoFile(outputFolder)
	csvFilePath := filepath.Join(outputFolder, "simulation_output.csv")
	videoFilePath := filepath.Join(outputFolder, "video.mp4")

	// Open a CSV file to record the infected states over time
	file, err := os.Create(csvFilePath)
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the CSV headers
	headers := []string{
		"Time", "virion_half_life", "dip_half_life", "ifn_half_life", "Global IFN Concentration Per Cell", "Total Extracellular Virions",
		"Total Extracellular DIPs", "Percentage Dead Cells", "Percentage Susceptible Cells",
		"Percentage Infected Cells", "Percentage Infected DIP-only Cells",
		"Percentage Infected Both Cells", "Percentage Antiviral Cells",
		"Regrowth Count",
		"Percentage Susceptible and Antiviral (Real Susceptible cells without regrowthed ones) Cells",
		"Percentage Regrowthed or Regrowthed and Antiviral Cells",
		"Probability Virion Infection", "Probability DIP Infection",
		"Per Particle Infection Chance RHO", "Total Local Particles",
		"Plaque Percentage", "max_global_IFN", "time_all_cells_uninfected",
		"Percentage Uninfected Cells", "num_plaques", "GRID_SIZE", "TIMESTEP",
		"IFN_DELAY", "STD_IFN_DELAY", "ALPHA", "RHO", "TAU", "BURST_SIZE_V",
		"REGROWTH_MEAN", "REGROWTH_STD", "TIME_STEPS", "MEAN_LYSIS_TIME",
		"STANDARD_LYSIS_TIME", "init_v_pfu_per_cell", "init_d_pfu_per_cell",
		"MEAN_ANTI_TIME_Per_Cell", "STD_ANTI_TIME", "R", "BURST_SIZE_D", "H",
		"option", "d_pfu_initial", "v_pfu_initial", "virionOnlyInfected", "dipOnlyInfected",
		"bothInfected", "totalDeadFromV", "totalDeadFromBoth", "virionDiffusionRate", "dipDiffusionRate", "k_JumpR",
		"jumpRadiusV", "jumpRadiusD", "jumpRandomly", "par_celltocell_random",
		"allowVirionJump", "allowDIPJump", "IFN_wave_radius", "ifnWave",
		"ifnBothFold", "D_only_IFN_stimulate_ratio", "BOTH_IFN_stimulate_ratio",
		"totalRandomJumpVirions", "totalRandomJumpDIPs", "dipAdvantage",
	}

	err = writer.Write(headers)
	if err != nil {
		log.Fatalf("Failed to write CSV headers: %v", err)
	}

	// Create an MJPEG video writer
	videoWriter, err := mjpeg.New(videoFilePath, int32(GRID_SIZE*CELL_SIZE*2), int32(GRID_SIZE*CELL_SIZE*2), int32(FRAME_RATE))
	if err != nil {
		log.Fatalf("Failed to create MJPEG writer: %v", err) // Handle the error if the writer fails to create
	}
	defer videoWriter.Close() // Ensure the writer is closed when the program ends

	var buf bytes.Buffer                       // Buffer for JPEG encoding
	jpegOptions := &jpeg.Options{Quality: 100} // JPEG encoding options, quality set to 75

	var frameNumbers []int            // Slice to store frame numbers
	var deadCellPercentages []float64 // Slice to store dead cell percentages
	virionOnly := make([]float64, TIME_STEPS)
	dipOnly := make([]float64, TIME_STEPS)
	both := make([]float64, TIME_STEPS)
	// Ensure the first frame has valid values
	virionOnly[0] = 0.0
	dipOnly[0] = 0.0
	both[0] = 0.0
	// Output image save directory

	var extractedImages []*image.RGBA // Store selected frame images

	for frameNum := 0; frameNum < TIME_STEPS; frameNum++ {

		grid.update(frameNum) // Update the grid state

		// Call the function to record infected state counts at the specific frames
		grid.recordSimulationData(writer, frameNum)

		// Calculate and record the percentage of dead cells, excluding regrowth cells
		deadCellsPercentage := calculateDeadCellPercentage(grid.state)
		frameNumbers = append(frameNumbers, frameNum)                          // Record the current frame number
		deadCellPercentages = append(deadCellPercentages, deadCellsPercentage) // Record the percentage of dead cells

		// Calculate infection percentages
		virionOnly[frameNum] = float64(grid.calculateVirionOnlyInfected()) / float64(GRID_SIZE*GRID_SIZE) * 100
		dipOnly[frameNum] = float64(grid.calculateDipOnlyInfected()) / float64(GRID_SIZE*GRID_SIZE) * 100
		both[frameNum] = float64(grid.calculateBothInfected()) / float64(GRID_SIZE*GRID_SIZE) * 100

		if frameNum > 1 {
			if frameNum%24 == 0 { // Save every 10 frames

				img := grid.gridToImageWithGraph(frameNum, virionOnly[:frameNum+1], dipOnly[:frameNum+1], both[:frameNum+1], videotype, false)

				extractedImages = append(extractedImages, img)
			}
		}

		// Log `y` values before feeding them to the graph
		log.Printf("Frame %d: Virion Only: %.2f%%, DIP Only: %.2f%%, Both: %.2f%%", frameNum, virionOnly[frameNum], dipOnly[frameNum], both[frameNum])
		// Generate the graph only if there are at least two frames of data
		var img *image.RGBA
		if frameNum > 0 {
			img = grid.gridToImageWithGraph(frameNum, virionOnly[:frameNum+1], dipOnly[:frameNum+1], both[:frameNum+1], videotype, true)
		} else {
			// For the first frame, only render the grid without the graph
			img = grid.gridToImage(videotype)
		}

		// Encode the image to JPEG format
		err = jpeg.Encode(&buf, img, jpegOptions)
		if err != nil {
			log.Fatalf("Failed to encode image: %v", err)
		}

		// Add the frame to the video
		err = videoWriter.AddFrame(buf.Bytes())
		if err != nil {
			log.Fatalf("Failed to add frame: %v", err)
		}
		buf.Reset() // Reset the buffer for the next frame

		if len(extractedImages) > 0 {
			combinedImage := combineImagesHorizontally(extractedImages)

			showLegend := false // ⬅️ Change here: control whether to add legend

			if showLegend {
				legendWidth := 180
				legendHeight := 80
				legendX := combinedImage.Bounds().Dx() - legendWidth - 20
				legendY := 20

				draw.Draw(combinedImage,
					image.Rect(legendX-5, legendY-5, legendX+legendWidth+5, legendY+legendHeight+5),
					&image.Uniform{color.RGBA{255, 255, 255, 200}},
					image.Point{}, draw.Over)

				addStaticLegend(combinedImage, legendX, legendY)
			}

			savePNGImage(combinedImage, filepath.Join(outputFolder, "selected_frames_combined.png"))
		}
	}
	log.Println("Video and graph saved successfully.") // Print a success message
	fmt.Println("ifnWave is ", ifnWave)
}
