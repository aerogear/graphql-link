//go:generate go run .

// This whole directory is used to generate the ../assets.go file.  It's not compiled into the final binary.
package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/chirino/hawtgo/sh"
	"github.com/shurcooL/httpfs/filter"
	"github.com/shurcooL/vfsgen"
)

func main() {
	projectDir, err := filepath.Abs(filepath.Join("..", "..", "..", "..", ".."))
	if err != nil {
		log.Fatalln(err)
	}

	YarnBuild(filepath.Join(projectDir, "ui"))

	err = vfsgen.Generate(GetAssetsFS(projectDir), vfsgen.Options{
		Filename:        filepath.Join("..", "assets.go"),
		PackageName:     "assets",
		BuildTags:       "",
		VariableName:    "FileSystem",
		VariableComment: "",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func YarnBuild(workingDir string) {
	sh.New().
		Dir(workingDir).
		CommandLog(os.Stdout).
		CommandLogPrefix("yarn > ").
		Line("yarn build").
		MustZeroExit()
}

func GetAssetsFS(projectDir string) http.FileSystem {
	assetsDir := filepath.Join(projectDir, "ui", "build")
	// to avoid changes due to changing timestamps...
	return NewFileInfoMappingFS(filter.Keep(http.Dir(assetsDir), func(path string, fi os.FileInfo) bool {
		if fi.Name() == ".DS_Store" {
			return false
		}
		//if strings.HasSuffix(fi.Name(), ".go") {
		//	return false
		//}
		return true
	}), func(fi os.FileInfo) (os.FileInfo, error) {
		return &zeroTimeFileInfo{fi}, nil
	})
}

type zeroTimeFileInfo struct {
	os.FileInfo
}

func (*zeroTimeFileInfo) ModTime() time.Time {
	return time.Time{}
}
