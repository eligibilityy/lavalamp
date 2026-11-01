package main

import (
	"math"
	"math/rand/v2"
)

const (
	buoyancy  = 0.2
	drag      = 1.5
	heatZone  = 0.25
	coolZone  = 0.25
	heatSpeed = 0.4
	midNudge  = 0.05
	sway      = 0.04
	stretch   = 2.0
	wobble    = 0.04
	maxStep   = 1.0 / 30
)

type blob struct {
	x, y     float64
	vx, vy   float64
	baseR    float64
	temp     float64
	heatRate float64
	phase    float64
	r        float64
	sx, sy   float64
}

type sim struct {
	w, h  float64
	t     float64
	size  float64
	blobs []blob
}

func newSim(w, h float64, count int, size float64) *sim {
	s := &sim{w: w, h: h, size: size}
	s.setCount(count)
	return s
}

func (s *sim) newBlob() blob {
	base := 0.07 + rand.Float64()*0.07
	r := base * s.size
	return blob{
		x:        r + rand.Float64()*math.Max(s.w-2*r, 0),
		y:        r + rand.Float64()*math.Max(s.h-2*r, 0),
		baseR:    base,
		temp:     rand.Float64(),
		heatRate: (0.1 / base) * (0.8 + rand.Float64()*0.4),
		phase:    rand.Float64() * 2 * math.Pi,
		r:        r,
		sx:       1,
		sy:       1,
	}
}

func (s *sim) setCount(n int) {
	for len(s.blobs) < n {
		s.blobs = append(s.blobs, s.newBlob())
	}
	s.blobs = s.blobs[:n]
}

func (s *sim) resize(w, h float64) {
	for i := range s.blobs {
		b := &s.blobs[i]
		b.x *= w / s.w
		b.y *= h / s.h
	}
	s.w, s.h = w, h
}

func (s *sim) advance(dt float64) {
	if dt <= 0 {
		return
	}
	n := int(math.Ceil(dt / maxStep))
	for range n {
		s.step(dt / float64(n))
	}
}

func (s *sim) step(dt float64) {
	s.t += dt
	for i := range s.blobs {
		b := &s.blobs[i]

		heat := smoothstep(1-heatZone, 1, (b.y+b.r)/s.h)
		cool := 1 - smoothstep(0, coolZone, (b.y-b.r)/s.h)
		nudge := midNudge * (0.5 - b.y/s.h)

		change := heatSpeed*heat*(1-b.temp) - heatSpeed*cool*b.temp + nudge
		b.temp = clamp(b.temp+change*b.heatRate*dt, 0, 1)

		ay := -buoyancy * s.h * (2*b.temp - 1)
		ax := sway * (math.Sin(s.t*0.31+b.phase) + 0.5*math.Sin(s.t*0.73+2*b.phase))

		b.vx += (ax - drag*b.vx) * dt
		b.vy += (ay - drag*b.vy) * dt

		b.x += b.vx * dt
		b.y += b.vy * dt

		b.r = b.baseR * s.size * (1 + wobble*math.Sin(s.t*1.3+3*b.phase))
		b.sy = math.Min(1+stretch*math.Abs(b.vy)/s.h, 1.5)
		b.sx = 1 / b.sy

		b.x, b.vx = bounce(b.x, b.vx, b.r*b.sx, s.w, 0.5)
		b.y, b.vy = bounce(b.y, b.vy, b.r*b.sy, s.h, 0)
	}
}

func bounce(pos, vel, r, size, restitution float64) (float64, float64) {
	if 2*r >= size {
		return size / 2, 0
	}
	if pos-r < 0 {
		return r, restitution * math.Abs(vel)
	}
	if pos+r > size {
		return size - r, -restitution * math.Abs(vel)
	}
	return pos, vel
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}
