package main

import (
	"time"
//	"math/rand"
	"math"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)


const BOARD_SIZE		= 8
const TILE_WIDTH		= 132.0
const TILE_HEIGHT		= 66.0

type CityGame_Tile struct {
	sprite	*pixel.Sprite
	details	*SpriteSheet_SubTexture
}

type CityGame_Elements struct {
	win							opengl.Window
	textAtlas					*text.Atlas

	// Text Labels
	fpsText						*text.Text
	scoreText					*text.Text					
	debugText					*text.Text
	mousePosText				*text.Text

	// Tile Data
	spritesheet					*pixel.Picture
	textureAtlas				*SpriteSheet_TextureAtlas
	tileBatch					*pixel.Batch
	
	// Current Tile Data
	boardTiles					[BOARD_SIZE][BOARD_SIZE]CityGame_Tile
	currentTileGroup			int
	currentTile					int

	// Display Tracking
	camPos						pixel.Vec
	camSpeed					float64
	camZoom						float64
	camZoomSpeed				float64

	frames						int
	second						<-chan time.Time
	debugMode					bool
	
	// Game State
	score						int
} 

func(el *CityGame_Elements) _generateTextElement(pos pixel.Vec) *text.Text {
	item := text.New(pos, el.textAtlas)
	item.Color = colornames.White
	return item
}

func(el *CityGame_Elements) _initTextElements() error {

	el.textAtlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)

	el.fpsText = el._generateTextElement(pixel.V(0, 200))
	el.scoreText = el._generateTextElement(pixel.V(0, 250))
	el.debugText = el._generateTextElement(pixel.V(0, 300))
	el.mousePosText = el._generateTextElement(pixel.V(0, 350))

	return nil
}

func(el *CityGame_Elements) _initTileElements() error {

	ss, err := loadPicture("resources/images/cityTiles_sheet.png")
	if err != nil {
		return err
	}
	el.spritesheet = &ss
	ta, err := loadTextureAtlas("resources/images/cityTiles_sheet.xml")
	if err != nil {
		return err
	}
	el.textureAtlas = &ta
	el.tileBatch = pixel.NewBatch(&pixel.TrianglesData{}, *el.spritesheet)

	return nil
}


func(el *CityGame_Elements) cycleNextTile() {
	el.currentTile++
	if el.currentTile >= len(el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures) {
		el.currentTile = 0
	}
}

func(el *CityGame_Elements) getNextTileGroup() {
	// Actual Game
	// el.currentTileGroup = rand.Intn(len(el.textureAtlas.TileGroups))
	// Debugging
	el.currentTileGroup++
	if el.currentTileGroup >= len(el.textureAtlas.TileGroups) {
		el.currentTileGroup = 0
	}

	el.currentTile = 0
}

func(el *CityGame_Elements) getCurrentTile() *CityGame_Tile {
	// Get the current tile bounds and create a sprite
	texture := el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile]
	tileBounds := pixel.R(
		(*el.spritesheet).Bounds().Min.X + float64(texture.X), 
		(*el.spritesheet).Bounds().Max.Y - float64(texture.Y) - float64(texture.Height), 
		(*el.spritesheet).Bounds().Min.X + float64(texture.X + texture.Width), 
		(*el.spritesheet).Bounds().Max.Y - float64(texture.Y))
	tileSprite := pixel.NewSprite(*el.spritesheet, tileBounds)
	tile := &CityGame_Tile{
		sprite:			tileSprite,
		details:		&el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile],
	}

	return tile
}

func(el *CityGame_Elements) getCurrentTileGroupName() string {
	return el.textureAtlas.TileGroups[el.currentTileGroup].Name
}

func(el *CityGame_Elements) getCurrentTileJoinTL() JoinType {
	return el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile].JoinTL
}

func(el *CityGame_Elements) getCurrentTileJoinTR() JoinType {
	return el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile].JoinTR
}

func(el *CityGame_Elements) getCurrentTileJoinBR() JoinType {
	return el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile].JoinBR
}

func(el *CityGame_Elements) getCurrentTileJoinBL() JoinType {
	return el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile].JoinBL
}

func (el *CityGame_Elements) checkAddCurrentTile (x int, y int, tile *CityGame_Tile) {

	if el.boardTiles[x][y].sprite == nil {
		el.boardTiles[x][y] = *tile

		// Score Checks!
		thisTile := &el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile]

		if x > 0 && el.boardTiles[x - 1][y].sprite != nil && 
			thisTile.JoinTL.compatibleWith(el.boardTiles[x - 1][y].details.JoinBR) {
			el.score++
		}
		if x < (BOARD_SIZE - 1) && el.boardTiles[x + 1][y].sprite != nil && 
			thisTile.JoinBR.compatibleWith(el.boardTiles[x + 1][y].details.JoinTL) {
			el.score++
		}
		if y > 0 && el.boardTiles[x][y - 1].sprite != nil && 
			thisTile.JoinBL.compatibleWith(el.boardTiles[x][y - 1].details.JoinTR) {
			el.score++
		}
		if y < (BOARD_SIZE - 1) && el.boardTiles[x][y + 1].sprite != nil && 
			thisTile.JoinTR.compatibleWith(el.boardTiles[x][y + 1].details.JoinBL) {
			el.score++
		}
	
		el.getNextTileGroup()
	}
}

func(el *CityGame_Elements) scrollLeft(dt float64) {
	el.camPos.X -= el.camSpeed * dt
}

func(el *CityGame_Elements) scrollRight(dt float64) {
	el.camPos.X += el.camSpeed * dt
}

func(el *CityGame_Elements) scrollUp(dt float64) {
	el.camPos.Y += el.camSpeed * dt
}

func(el *CityGame_Elements) scrollDown(dt float64) {
	el.camPos.Y -= el.camSpeed * dt
}

func(el *CityGame_Elements) zoom(zoomLevel float64) {
	el.camZoom *= math.Pow(el.camZoomSpeed, zoomLevel)
}



func InitGameElements() (*CityGame_Elements, error) {
	e := &CityGame_Elements{
		currentTileGroup: 	0,
		currentTile: 		0,
		camPos: 			pixel.V(480.0, 120.0),
		camSpeed: 			500.0,
		camZoom:			1.0,
		camZoomSpeed:		1.2,
		frames:				0,
		second:				time.Tick(time.Second),
		debugMode:			false,

		score:				0,
	}

	err := e._initTextElements()
	if err != nil {
		return nil, err
	}

	err = e._initTileElements()
	if err != nil {
		return nil, err
	}

	return e, nil
}
