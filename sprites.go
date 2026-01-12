package main

import (
	"os"
	"io/ioutil"
	"encoding/xml"
)

/*******
 * JoinType Definitions
 ********/
type JoinType int

const (
	JoinType_Concrete 			= 0
	JoinType_Road				= 1
	JoinType_RoadFootpathRight	= 2
	JoinType_RoadFootpathLeft	= 3
	JoinType_RoadFullWidth		= 4
	JoinType_Canal				= 5
	JoinType_BikePath			= 6
	JoinType_NatureStrip		= 7
	JoinType_TunnelHeight		= 8
	JoinType_TunnelDepth		= 9
	JoinType_RedWallFull		= 10
	JoinType_RedWallGap			= 11
	JoinType_RoadHeight			= 12
	JoinType_River				= 13
)

var joinTypeName = map[JoinType]string {
	JoinType_Concrete 			: "Concrete",
	JoinType_Road				: "Road",
	JoinType_RoadFootpathRight	: "Road With Footpath On Right",
	JoinType_RoadFootpathLeft	: "Road With Footpath On Left",
	JoinType_RoadFullWidth		: "Full Width Road",
	JoinType_Canal				: "Canal",
	JoinType_BikePath			: "Bike Path",
	JoinType_NatureStrip		: "Nature Strip",
	JoinType_TunnelHeight		: "Tunnel Height",
	JoinType_TunnelDepth		: "Tunnel Depth",
	JoinType_RedWallFull		: "Red Wall",
	JoinType_RedWallGap			: "Red Wall With Gap",
	JoinType_RoadHeight			: "Road Height",
	JoinType_River				: "River",
}

func (jt JoinType) String() string {
	return joinTypeName[jt]
}

func (jt JoinType) compatibleWith(comp JoinType) bool {
	// Concrete gets nothing
	if jt == JoinType_Concrete {
		return false
	} else if (jt == JoinType_RoadFootpathLeft && comp == JoinType_RoadFootpathRight) ||
			(jt == JoinType_RoadFootpathRight && comp == JoinType_RoadFootpathLeft) {
		// Footpath bits go opposite I guess?
		return true
	}
	return jt == comp
}

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
	JoinTL  JoinType	`xml:"joinTL,attr"`
	JoinTR  JoinType	`xml:"joinTR,attr"`
	JoinBR  JoinType	`xml:"joinBR,attr"`
	JoinBL  JoinType	`xml:"joinBL,attr"`
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
