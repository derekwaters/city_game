package main

import (
	"time"
	"math/rand"
	"math"
	"fmt"
	"image/color"

	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"golang.org/x/image/colornames"
)


const BOARD_SIZE		= 8
const TILE_WIDTH		= 132.0
const TILE_HEIGHT		= 66.0
const ALPHA_FADE_AMOUNT	= 2

type GameState 		int
type MenuSelection 	int

const (
	GameState_Title = iota
	GameState_Settings
	GameState_About
	GameState_ReallyQuit
	GameState_Running
	GameState_GameOver
)

const (
	MenuSelection_NewGame = iota
	MenuSelection_Settings
	MenuSelection_About
	MenuSelection_Quit
)

var MenuSelectionName = map[MenuSelection]string{
	MenuSelection_NewGame:	"New Game",
	MenuSelection_Settings:	"Settings",
	MenuSelection_About:	"About",
	MenuSelection_Quit:		"Quit",
}

type CityGame_Tile struct {
	sprite				*pixel.Sprite
	details				*SpriteSheet_SubTexture
	highlightColor		color.RGBA
	highlightTile		bool
}

type CityGame_Elements struct {
	win							opengl.Window
	textAtlas					*text.Atlas
	debugAtlas					*text.Atlas

	// Polygon Drawing
	imdraw						*imdraw.IMDraw

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
	state						GameState			
	menuSelection				MenuSelection
} 

func(el *CityGame_Elements) _generateTextElement(pos pixel.Vec, face *text.Atlas) *text.Text {
	item := text.New(pos, face)
	item.Color = colornames.White
	return item
}

func(el *CityGame_Elements) _initTextElements() error {

	face, err := loadTTF("resources/fonts/Chibi.ttf", 40)
	if err != nil {
		panic(err)
	}

	el.textAtlas = text.NewAtlas(face, text.ASCII)

	debugFace, err := loadTTF("resources/fonts/Chibi.ttf", 15)
	if err != nil {
		panic(err)
	}

	el.debugAtlas = text.NewAtlas(debugFace, text.ASCII)


	el.fpsText = el._generateTextElement(pixel.V(-60, 250), el.debugAtlas)
	el.scoreText = el._generateTextElement(pixel.V(-60, 400), el.textAtlas)
	el.debugText = el._generateTextElement(pixel.V(-60, 350), el.debugAtlas)
	el.mousePosText = el._generateTextElement(pixel.V(-60, 300), el.debugAtlas)

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

func(el *CityGame_Elements) _initDrawingElements() {
	el.imdraw = imdraw.New(nil)

	el.imdraw.Color = pixel.RGB(0.0, 0.2, 0.0)
	for x := 0; x <= BOARD_SIZE; x++ {
		p1 := el.mapTilePos(x, 0).Sub(pixel.V(TILE_WIDTH / 2.0, -TILE_HEIGHT / 2.0))
		p2 := el.mapTilePos(x, BOARD_SIZE).Sub(pixel.V(TILE_WIDTH / 2.0, -TILE_HEIGHT / 2.0))
		el.imdraw.Push(p1)
		el.imdraw.Push(p2)
		el.imdraw.Line(3)
	} 

	for y := 0; y <= BOARD_SIZE; y++ {
		p1 := el.mapTilePos(0, y).Sub(pixel.V(TILE_WIDTH / 2.0, -TILE_HEIGHT / 2.0))
		p2 := el.mapTilePos(BOARD_SIZE, y).Sub(pixel.V(TILE_WIDTH / 2.0, -TILE_HEIGHT / 2.0))
		el.imdraw.Push(p1)
		el.imdraw.Push(p2)
		el.imdraw.Line(3)
	} 
}

func(el *CityGame_Elements) mapTilePos(x int, y int) pixel.Vec {
	return pixel.V(
		(float64(x) * TILE_WIDTH / 2.0) + (float64(y) * TILE_WIDTH / 2.0),
		(-1.0 * float64(x) * TILE_HEIGHT / 2.0) + (float64(y) * TILE_HEIGHT / 2.0))
}

func(el *CityGame_Elements) cycleNextTile() {
	el.currentTile++
	if el.currentTile >= len(el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures) {
		el.currentTile = 0
	}
}

func(el *CityGame_Elements) getNextTileGroup() {
	// Actual Game
	el.currentTileGroup = rand.Intn(len(el.textureAtlas.TileGroups))
	// Debugging
	// el.currentTileGroup++
	// if el.currentTileGroup >= len(el.textureAtlas.TileGroups) {
	// 	el.currentTileGroup = 0
	// }

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

func(el *CityGame_Elements) getTotalTileCount() int {
	ret := 0
	for x := 0; x < BOARD_SIZE; x++ {
		for y := 0; y < BOARD_SIZE; y++ {
			if el.boardTiles[x][y].sprite != nil {
				ret++
			}
		}
	}
	return ret
}

func (el *CityGame_Elements) checkAddCurrentTile (x int, y int, tile *CityGame_Tile) {

	if el.boardTiles[x][y].sprite == nil {
		el.boardTiles[x][y] = *tile

		// Score Checks!
		thisTile := &el.textureAtlas.TileGroups[el.currentTileGroup].SubTextures[el.currentTile]

		if (x > 0 && el.boardTiles[x - 1][y].sprite != nil && 
			thisTile.JoinTL.compatibleWith(el.boardTiles[x - 1][y].details.JoinBR)) ||
		   (x < (BOARD_SIZE - 1) && el.boardTiles[x + 1][y].sprite != nil && 
			thisTile.JoinBR.compatibleWith(el.boardTiles[x + 1][y].details.JoinTL)) ||
		   (y > 0 && el.boardTiles[x][y - 1].sprite != nil && 
			thisTile.JoinBL.compatibleWith(el.boardTiles[x][y - 1].details.JoinTR)) ||
		   (y < (BOARD_SIZE - 1) && el.boardTiles[x][y + 1].sprite != nil && 
			thisTile.JoinTR.compatibleWith(el.boardTiles[x][y + 1].details.JoinBL)) {
			el.score++
			el.boardTiles[x][y].highlightTile = true
			el.boardTiles[x][y].highlightColor = colornames.Yellow
			el.boardTiles[x][y].highlightColor.A = 0
		}
	
		if el.getTotalTileCount() == (BOARD_SIZE * BOARD_SIZE) {
			panic(fmt.Sprintf("GAME OVER: Score = %d", el.score))
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

func(el *CityGame_Elements) resetGame() {
	el.score = 0
	for x := 0; x < BOARD_SIZE; x++ {
		for y := 0; y < BOARD_SIZE; y++ {
			el.boardTiles[x][y].details = nil
			el.boardTiles[x][y].sprite = nil
		}
	}
	el.getNextTileGroup()
}


func(el *CityGame_Elements) doFades() {
	for x := 0; x < BOARD_SIZE; x++ {
		for y := 0; y < BOARD_SIZE; y++ {
			if el.boardTiles[x][y].highlightTile {
				if el.boardTiles[x][y].highlightColor.A < (0xFF - ALPHA_FADE_AMOUNT) {
					el.boardTiles[x][y].highlightColor.A += ALPHA_FADE_AMOUNT
				} else {
					el.boardTiles[x][y].highlightTile = false
				}
			}
		}
	}
}

func InitGameElements() (*CityGame_Elements, error) {
	e := &CityGame_Elements{
		state:				GameState_Title,
		menuSelection:		MenuSelection_NewGame,
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

	e._initDrawingElements()

	e.getNextTileGroup()
	
	return e, nil
}
