package drones

import (
	"math/rand"
	"strings"

	"github.com/itchyny/maze"

	"github.com/bot-games/drones/pb"
)

const (
	width  = 5
	height = 11
)

var format = &maze.Format{
	Wall:               "█",
	Path:               " ",
	StartLeft:          " ",
	StartRight:         " ",
	GoalLeft:           " ",
	GoalRight:          " ",
	Solution:           " ",
	SolutionStartLeft:  " ",
	SolutionStartRight: " ",
	SolutionGoalLeft:   " ",
	SolutionGoalRight:  " ",
	Visited:            " ",
	VisitedStartLeft:   " ",
	VisitedStartRight:  " ",
	VisitedGoalLeft:    " ",
	VisitedGoalRight:   " ",
	Cursor:             " ",
}

type Maze struct {
	Width  uint8
	Height uint8
	Walls  []byte
}

type Position struct {
	X, Y uint8
}

func NewCheckPoints(maze *Maze) []*pb.Options_CellPos {
	result := make([]*pb.Options_CellPos, 5)

	var available []Position
	for i, zone := range []struct {
		X1, Y1, X2, Y2 uint8
	}{
		{1, 1, maze.Width / 2, maze.Height / 2},
		{1, maze.Height/2 + 1, maze.Width / 2, maze.Height - 1},
		{maze.Width/2 + 1, maze.Height/2 + 1, maze.Width - 1, maze.Height - 1},
		{maze.Width/2 + 1, 1, maze.Width - 1, maze.Height / 2},
	} {
		available = available[:0]
		for y := zone.Y1; y <= zone.Y2; y++ {
			for x := zone.X1; x <= zone.X2; x++ {
				if !maze.IsWall(x, y) {
					available = append(available, Position{x, y})
				}
			}
		}
		pos := available[rand.Intn(len(available))]
		result[i] = &pb.Options_CellPos{
			X: uint32(pos.X),
			Y: uint32(pos.Y),
		}
	}

	result[4] = &pb.Options_CellPos{
		X: 22,
		Y: 1,
	}

	return result
}

func NewMaze() *Maze {
	var directions1 [][]int
	for x := 0; x < height; x++ {
		directions1 = append(directions1, make([]int, width))
	}

	var directions2 [][]int
	for x := 0; x < height; x++ {
		directions2 = append(directions2, make([]int, width))
	}

	m1 := &maze.Maze{directions1, height, width,
		&maze.Point{height - 1, 0}, &maze.Point{0, width - 1}, &maze.Point{height - 1, 0},
		false, false, false}

	m2 := &maze.Maze{directions2, height, width,
		&maze.Point{0, 0}, &maze.Point{height - 1, width - 1}, &maze.Point{0, 0},
		false, false, false}

	m1.Generate()
	m2.Generate()

	str1 := strings.Split(m1.String(format), "\n")[1:]
	str2 := strings.Split(m2.String(format), "\n")[1:]

	res := &Maze{
		Width:  24,
		Height: 23,
		Walls:  make([]byte, (len(str1)-2)*(len([]rune(str1[0]))+len([]rune(str2[0])))/8),
	}

	for y := 0; y < len(str1)-2; y++ {
		res.setWall(0, uint8(y))
		for x, r := range []rune(str1[y][1:]) {
			if r == '█' {
				res.setWall(uint8(x+1), uint8(y))
			}
		}

		offset := len([]rune(str1[y][1:]))
		for x, r := range []rune(str2[y][1:]) {
			if r == '█' {
				res.setWall(uint8(offset+x+1), uint8(y))
			}
		}

		res.setWall(uint8(offset+offset+1), uint8(y))
	}

	return res
}

func (m *Maze) IsWall(x, y uint8) bool {
	i := int(m.Height-y-1)*int(m.Width) + int(x)
	byteIndex := i / 8
	bitIndex := uint(i % 8)
	return m.Walls[byteIndex]&(1<<bitIndex) > 0
}

func (m *Maze) String() string {
	var res string
	for y := uint8(0); y < m.Height; y++ {
		for x := uint8(0); x < m.Width; x++ {
			if m.IsWall(x, m.Height-y-1) {
				res += "█"
			} else {
				res += " "
			}
		}
		res += "\n"
	}

	return res
}

func (m *Maze) Solve(start, goal Position) []Position {
	visited := make(map[Position]bool)
	path := make([]Position, 0)
	m.solveRecursive(goal, start, visited, &path)
	return path
}

func (m *Maze) solveRecursive(goal, currentPos Position, visited map[Position]bool, path *[]Position) bool {
	if currentPos == goal {
		*path = append(*path, currentPos)
		return true
	}

	if m.IsWall(currentPos.X, currentPos.Y) || visited[currentPos] {
		return false
	}

	visited[currentPos] = true
	*path = append(*path, currentPos)

	for _, direction := range []struct {
		X int8
		Y int8
	}{
		{0, -1},
		{0, 1},
		{-1, 0},
		{1, 0},
	} {
		if currentPos.X == 0 && direction.X < 0 ||
			currentPos.Y == 0 && direction.Y < 0 {
			continue
		}
		newPos := Position{
			X: uint8(int8(currentPos.X) + direction.X),
			Y: uint8(int8(currentPos.Y) + direction.Y),
		}

		if m.solveRecursive(goal, newPos, visited, path) {
			return true
		}
	}

	*path = (*path)[:len(*path)-1]
	delete(visited, currentPos)
	return false
}

func (m *Maze) setWall(x, y uint8) {
	i := int(y)*int(m.Width) + int(x)
	byteIndex := i / 8
	bitIndex := uint(i % 8)
	m.Walls[byteIndex] |= 1 << bitIndex
}
