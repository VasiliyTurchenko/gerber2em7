package configurator

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
)

const (
	CfgCommonPrintMemoryInfo     string = "common.PrintMemoryInfo"
	CfgCommonPrintAperturesInfo  string = "common.PrintAperturesInfo"
	CfgCommonPrintRegionsInfo    string = "common.PrintRegionsInfo"
	CfgCommonPrintStatistic      string = "common.PrintStatistic"
	CfgParserSaveIntermediate    string = "parser.SaveIntermediate"
	CfgCommonPrintGerberComments string = "common.PrintGerberComments"
	CfgRendererOutFile           string = "renderer.OutFile"
	CfgRendererGeneratePNG       string = "renderer.GeneratePNG"

	CfgPlotterOutFile  string = "plotter.OutFile"
	CfgPlotterXRes     string = "plotter.xRes"
	CfgPlotterYRes     string = "plotter.yRes"
	CfgPlotterPenSizes string = "plotter.PenSizes"
)

const (
	CfgRenderDrawContours    string = "renderer.DrawContours"
	CfgRenderDrawMoves       string = "renderer.DrawMoves"
	CfgRenderDrawOnlyRegions string = "renderer.DrawOnlyRegions"
	CfgPrintRegionInfo       string = "renderer.PrintRegionInfo"
)

const (
	CfgFoldersPlotterFilesFolder string = "folders.PlotterFilesFolder"
	CfgFoldersIntermediateFilesFolder string = "folders.IntermediateFilesFolder"
	CfgFoldersPNGFilesFolder string  ="folders.PNGFilesFolder"
)

func SetDefaults(v *viper.Viper) {
	v.SetConfigName("config") // no need to include file extension
	v.AddConfigPath(".")      // set the path of your config file
	v.SetConfigType("toml")

	// diagnostic messages
	v.SetDefault(CfgCommonPrintMemoryInfo, true)
	v.SetDefault(CfgCommonPrintAperturesInfo, true)
	v.SetDefault(CfgCommonPrintRegionsInfo, true)
	v.SetDefault(CfgCommonPrintStatistic, true)
	v.SetDefault(CfgCommonPrintGerberComments, true)

	//
	v.SetDefault(CfgParserSaveIntermediate, true)
	v.SetDefault(CfgRendererGeneratePNG, true)
	v.SetDefault(CfgRendererOutFile, "out.png")

	//
	v.SetDefault("pcb.xOrigin", 0)
	v.SetDefault("pcb.yOrigin", 0)

	//
	v.SetDefault("renderer.CanvasWidth", 297)
	v.SetDefault("renderer.CanvasHeight", 210)

	v.SetDefault(CfgRenderDrawContours, false)
	v.SetDefault(CfgRenderDrawMoves, false)
	v.SetDefault(CfgRenderDrawOnlyRegions, false)
	v.SetDefault(CfgPrintRegionInfo, false)

	// TODO: something wrong with the default values
	v.SetDefault(CfgPlotterPenSizes, []float64{0.07, 0.07, 0.07, 0.00})
	v.SetDefault(CfgPlotterXRes, 0.025)
	v.SetDefault(CfgPlotterYRes, 0.025)
	v.SetDefault(CfgPlotterOutFile, "plotter.out")

/*
[folders]
PlotterFilesFolder = "G:\go_prj\gerber2em7\plt"
IntermediateFilesFolder = "G:\go_prj\gerber2em7\tmp"
PNGFilesFolder = "G:\go_prj\gerber2em7\png"
 */
	v.SetDefault(CfgFoldersPlotterFilesFolder, "")
	v.SetDefault(CfgFoldersIntermediateFilesFolder, "")
	v.SetDefault(CfgFoldersPNGFilesFolder, "")
}

func ProcessConfigFile(v *viper.Viper) error {
	return v.ReadInConfig()
	return errors.New("configuration file error. Using defaults")
}

func DiagnosticAllCfgPrint(v *viper.Viper) {
	c := v.AllSettings()
	for key, data := range c {
		fmt.Println(key, ":", data)
	}
	fmt.Println()
	return
}
