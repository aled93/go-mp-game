package systems

import (
	"gomp/examples/new-api/components"
	"gomp/examples/new-api/config"
	"gomp/pkg/ecs"
	"gomp/pkg/kbd"
	"gomp/pkg/render"
	"gomp/pkg/util"
	"gomp/stdcomponents"
	"image/color"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func NewMinimapSystem() MinimapSystem {
	return MinimapSystem{}
}

type MinimapSystem struct {
	EntityManager *ecs.EntityManager
	Cameras       *stdcomponents.CameraComponentManager
	Position      *stdcomponents.PositionComponentManager
	Rotation      *stdcomponents.RotationComponentManager
	RlTexturePros *stdcomponents.RLTextureProComponentManager
	Renderables   *stdcomponents.RenderableComponentManager
	FrameBuffer2D *stdcomponents.FrameBuffer2DComponentManager
	Player        *components.PlayerTagComponentManager

	minimapCamera        ecs.Entity
	frameBufferComponent stdcomponents.FrameBuffer2D
	cameraComponent      stdcomponents.Camera
	disabled             bool
}

func (s *MinimapSystem) Init() {
	width, height := rl.GetScreenWidth(), rl.GetScreenHeight()

	s.minimapCamera = s.EntityManager.Create()
	s.Cameras.Create(s.minimapCamera, stdcomponents.Camera{
		Target:   util.Vec2{},
		Offset:   util.NewVec2(width, height).ScaleScalar(0.5),
		Rotation: 0,
		Zoom:     .5,
		Dst: util.NewRectFromOriginSize(
			util.NewVec2(0, width-height*(5/3)),
			util.NewVec2(width*(5/3), height*(5/3)),
		),
		Layer:     config.MinimapCameraLayer,
		Order:     1,
		Culling:   stdcomponents.Culling2DFullscreenBB,
		BlendMode: render.BlendAlpha,
		BGColor:   color.RGBA{R: 0, G: 0, B: 0, A: 255},
		Tint:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
	})
	s.FrameBuffer2D.Create(s.minimapCamera, stdcomponents.FrameBuffer2D{
		Position:  util.Vec2{},
		Frame:     util.NewRect(0, 0, width, height),
		Texture:   rl.LoadRenderTexture(int32(width), int32(height)),
		Layer:     config.MinimapCameraLayer,
		BlendMode: render.BlendAlpha,
		Rotation:  0,
		Tint:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Dst: util.NewRectFromOriginSize(
			util.NewVec2(0, height-height*(5/3)),
			util.NewVec2(width*(5/3), height*(5/3)),
		),
	})
}

func (s *MinimapSystem) Run(dt time.Duration) bool {
	c := s.Cameras.GetUnsafe(s.minimapCamera)

	if kbd.IsKeyPressed(kbd.KeycodeM) {
		if s.disabled {
			s.disabled = false
			c = s.Cameras.Create(s.minimapCamera, s.cameraComponent)
			s.FrameBuffer2D.Create(s.minimapCamera, s.frameBufferComponent)
		} else {
			s.disabled = true
			s.cameraComponent = *c
			s.frameBufferComponent = *s.FrameBuffer2D.GetUnsafe(s.minimapCamera)
			s.Cameras.Delete(s.minimapCamera)
			s.FrameBuffer2D.Delete(s.minimapCamera)
		}
	}
	if s.disabled {
		return false
	}
	s.Player.EachEntity()(func(entity ecs.Entity) bool {
		playerPosition := s.Position.GetUnsafe(entity)
		rotation := s.Rotation.GetUnsafe(entity)

		c.Target.X = playerPosition.XY.X
		c.Target.Y = playerPosition.XY.Y
		c.Rotation = -rotation.Degrees()
		return false
	})
	return false
}

func (s *MinimapSystem) Destroy() {
}
