package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Item struct {
	Name string
	URL  string
}

func main3() {
	items := []Item{
		{Name: "Cafe Latte", URL: "https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Ftse4.mm.bing.net%2Fth%2Fid%2FOIP.84Xiur00njuWhv3z8__VeQAAAA%3Fcb%3Ducfimg2%26pid%3DApi%26ucfimg%3D1&f=1&ipt=8c2f22e18d1e5ef83365db2bd940a1cb26c222a36f2697acc82f10facc31ce17&ipo=images"},
		{Name: "Chocolate", URL: "https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Ftse2.explicit.bing.net%2Fth%2Fid%2FOIP.oPCLBsM0HmuNhYmUfUm6KQHaIB%3Fpid%3DApi&f=1&ipt=d0e8c1b963e7f7303d9cca186a3dbdbdd6d4180085168b5461b51dd189970753&ipo=images"},
		{Name: "Matcha", URL: "https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Ftse2.mm.bing.net%2Fth%2Fid%2FOIP.kVSNsUysQf6judNk7xDikAHaHa%3Fcb%3Ducfimg2%26pid%3DApi%26ucfimg%3D1&f=1&ipt=5fd42f27ce6e115dc7cf4f0fa74db35b4dadcf6eb8e5245b42a2967d67b18168&ipo=images"},
		{Name: "Americano", URL: "https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Ftse4.mm.bing.net%2Fth%2Fid%2FOIP.3ZsnJZR0MbumNbz-MZ4o_gHaFJ%3Fcb%3Ducfimg2%26pid%3DApi%26ucfimg%3D1&f=1&ipt=d31878be1364384ae2202995a6b9efdc7753c5ae21181b3ed5fc2bfe8b2776b8&ipo=images"},
		{Name: "Rose Latte", URL: "https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Ftse3.mm.bing.net%2Fth%2Fid%2FOIP.F1xv57mlfPe-IX8dOJhEeAHaJ4%3Fpid%3DApi&f=1&ipt=1c5747e8c204c712521275ac699cbc6b7aa89b4b782c673f88e90732c28952fb&ipo=images"},
		{Name: "Ice Lemon Tea", URL: "https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Ftse4.mm.bing.net%2Fth%2Fid%2FOIP.FTmJuVPt8vtQ5g6Vlrxd2QHaLH%3Fcb%3Ducfimg2%26pid%3DApi%26ucfimg%3D1&f=1&ipt=4d5f8bb83ff783b0967dd174e176963b982884752c56cfd1719980d62c687242&ipo=images"},
		{Name: "Sirap", URL: "https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Ftse3.mm.bing.net%2Fth%2Fid%2FOIP.fqoEC5zODLT9Q1Ngx_GrKgHaHa%3Fcb%3Ducfimg2%26pid%3DApi%26ucfimg%3D1&f=1&ipt=98f443ed3e447f51cb2efcbdd6af81618b36dab987677e512adc2cdc7b43fee7&ipo=images"},
		{Name: "Can Drinks", URL: "https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Ftse4.mm.bing.net%2Fth%2Fid%2FOIP.3OlZB-1BEDnYg_gkNNXyWQHaHa%3Fcb%3Ducfimg2%26pid%3DApi%26ucfimg%3D1&f=1&ipt=f46154a9050da7f33731fe38b7a30be97f51eac8beb9643e2854398802d6d807&ipo=images"},
		{Name: "Mineral Water", URL: "https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Ftse4.mm.bing.net%2Fth%2Fid%2FOIP.udt7FoLmJXkZty4pSo-7JwHaE8%3Fcb%3Ducfimgc2%26pid%3DApi&f=1&ipt=1a457db9ee9494ce3e2ede1748a2127d3ccaa1cc81251e8e331b9ea5b636481d&ipo=images"},
	}

	// Create directory
	os.MkdirAll("./images", os.ModePerm)

	for _, item := range items {
		if item.URL == "" {
			continue
		}

		// Sanitize filename: lowercase, replace spaces and special chars with underscores
		reg := regexp.MustCompile("[^a-z0-9]+")
		filename := reg.ReplaceAllString(strings.ToLower(item.Name), "_")
		filename = strings.Trim(filename, "_") + ".webp"
		filepath := "./images/" + filename

		fmt.Printf("Downloading %s...\n", filename)
		err := downloadFile(filepath, item.URL)
		if err != nil {
			fmt.Printf("Error downloading %s: %v\n", item.Name, err)
		}
	}
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
