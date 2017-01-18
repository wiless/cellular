package deployment

import (
	"log"

	"github.com/wiless/vlib"
)

func HexWrapGrid(N int, center vlib.Location3D, hexsize float64, RDEGREE float64, TrueCells int) (pts []vlib.Location3D, vCellIDs vlib.VectorI) {
	directions := []vlib.Location3D{{1, -1, 0}, {1, 0, -1}, {0, +1, -1}, {-1, +1, 0}, {-1, 0, +1}, {0, -1, +1}}
	// origin := vlib.Origin3D
	result := make([]vlib.Location3D, N)
	Mirrors := CubeMirrors(2)
	vCellIDs.Resize(N)
	LookUpCellID := make(map[vlib.Location3D]int)
	// cnt := 0
	n := 1
	breakloop := true
	if N > 1 {
		breakloop = false
	}
	ROTATE := 0
	if RDEGREE > 0 {
		ROTATE = 1
	}

	for r := 0; !breakloop; r++ {
		radius := float64(r)
		cube := directions[4+ROTATE].Scale3D(radius)
		for i := 0; i < 6; i++ {
			for j := 0; j < r; j++ {

				result[n] = Cube2XY(cube, hexsize)
				KK := i + ROTATE
				if KK == 6 {
					KK = 0
				}
				cube = directions[KK].Shift3D(cube)
				// log.Printf("Index Cubes %d : %v", n, cube)
				if n < TrueCells {

					LookUpCellID[cube] = n
					vCellIDs[n] = n
				}
				//cnt++
				if n >= TrueCells {

					/// Loading virtual cells

					d, idx, delta := ClosestMirror(Mirrors, cube)
					vid, ok := LookUpCellID[delta]
					_, _ = d, idx
					vCellIDs[n] = vid
					// log.Printf(" n=%d Closest Mirror is %d @ %v: distance %v, DELTA : %v , VID = %v", n, idx, Mirrors[idx], d, delta, vid)

					if !ok {
						log.Panic("Unable to locate Mirror ", cube, "Exra info ", d, idx, delta)
					}
				}
				// fmt.Printf("\n%d %d %v", n, vCellIDs[n], result[n].Cmplx())

				n = n + 1

				if n >= N {
					breakloop = true
					break
				}
			}

			if breakloop {

				break
			}
		}
	}
	for indx, res := range result {
		if indx != 0 {
			result[indx] = vlib.FromCmplx(res.Cmplx()*vlib.GetEJtheta(RDEGREE) + center.Cmplx())
		}
	}

	// log.Print Mirror centers
	// log.Println(result)
	//log.Printf("\nAll results grid points %f ", result)
	return result, vCellIDs
}

func Cube2XY(cube vlib.Location3D, hexsize float64) vlib.Location3D {
	var result vlib.Location3D
	x := hexsize * 1.7320508 * (cube.X + cube.Z*0.5) // sqrt(3)=1.7320508
	y := hexsize * 1.5 * cube.Z
	result.X, result.Y = y, x
	return result
}

func ClosestMirror(mirrorTable []vlib.Location3D, pt vlib.Location3D) (distance float64, indx int, dv vlib.Location3D) {
	distance = 15000
	indx = -1

	for i, c := range mirrorTable {
		src := Cube2XY(pt, 1)
		dest := Cube2XY(c, 1)
		dc := dest.DistanceFrom(src)
		// dc := (math.Abs(pt.X-c.X) + math.Abs(pt.Y-c.Y) + math.Abs(pt.Z-c.Z)) / 2.0
		// dc := pt.DistanceFrom(c)

		// log.Println("Index %d", i, c, dc)
		if distance > dc {
			distance = dc
			indx = i
			dv.SetXYZ(pt.X-c.X, pt.Y-c.Y, pt.Z-c.Z)
		}
	}
	return distance, indx, dv
}

func CubeMirrors(r int) []vlib.Location3D {
	var radius float64 = float64(r)
	// directions := []vlib.Location3D{{1, -1, 0}, {1, 0, -1}, {0, +1, -1}, {-1, +1, 0}, {-1, 0, +1}, {0, -1, +1}}
	oldcenter := vlib.Location3D{2.0*radius + 1, -radius, -radius - 1}
	mirrorTables := make([]vlib.Location3D, 6)
	mirrorTables[0] = oldcenter
	var newcenter vlib.Location3D

	for i := 1; i < 6; i++ {

		newcenter.SetXYZ(-oldcenter.Y, -oldcenter.Z, -oldcenter.X)
		mirrorTables[i] = newcenter
		oldcenter.SetXYZ(newcenter.X, newcenter.Y, newcenter.Z)
	}
	// log.Println("Mirror Table", mirrorTables)
	return mirrorTables
	// directions = [[1, -1, 0]; [1, 0, -1]; [0, +1, -1]; [-1, +1, 0]; [-1, 0, +1]; [0, -1, +1]]
	// FINALRADIUS=2;
	// mirrorCenter= [2*FINALRADIUS+1, -FINALRADIUS, -FINALRADIUS-1]

	// mirrorTables(1,:)=mirrorCenter;
	// oldcenter=mirrorCenter;
	// for k=2:6
	//     newcenter=-oldcenter;
	//     newcenter=[oldcenter(end) oldcenter(1:end-1)];
	//     mirrorTables(k,:)=newcenter;
	//     oldcenter=newcenter;
	// end
}
