package main

import (
	"os"
	"io/ioutil"
	"encoding/xml"
)

/*******
 * ImageMap Definitions
 ********/
type SpriteSheet_SubTexture struct {
	XMLName xml.Name 	`xml:"SubTexture"`
	Name	string		`xml:"name,attr"`
	X		int			`xml:"x,attr"`
	Y		int			`xml:"y,attr"`
	Width	int			`xml:"width,attr"`
	Height	int			`xml:"height,attr"`
} 

type SpriteSheet_TileGroup struct {
	XMLName		xml.Name					`xml:"TileGroup"`
	Name		string						`xml:"name,attr"`
	SubTextures	[]SpriteSheet_SubTexture	`xml:"SubTexture"`
}

type SpriteSheet_TextureAtlas struct {
	XMLName		xml.Name					`xml:"TextureAtlas"`
	TileGroups	[]SpriteSheet_TileGroup		`xml:"TileGroup"`
}

func loadTextureAtlas(path string) (SpriteSheet_TextureAtlas, error) {

	var atlas SpriteSheet_TextureAtlas

	file, err := os.Open(path)
	if err != nil {
		return atlas, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return atlas, err
	}

	err = xml.Unmarshal(bytes, &atlas)
	if err != nil {
		return atlas, err
	}

	return atlas, nil
}
