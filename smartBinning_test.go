/* Copyright (C) 2016 Philipp Benner
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package smartBinning

/* -------------------------------------------------------------------------- */

//import   "fmt"
import   "testing"

/* -------------------------------------------------------------------------- */

func Test1(t *testing.T) {

  x := []float64{-100,-99,1,2,3,6,8,19,21,120,300,350,355,380}
  y := []float64{1,2,3,4,5,6,7,8,9,10,11,12,13}

  binning, _ := New(x, y, BinSum, BinLessSize)

  if binning.Smallest.Lower != 1 {
    t.Error("test failed")
  }
  if binning.Smallest.Upper != 2 {
    t.Error("test failed")
  }
  if binning.Largest.Lower != 120 {
    t.Error("test failed")
  }
  if binning.Largest.Upper != 300 {
    t.Error("test failed")
  }

  // reduce number of bins
  binning.FilterBins(5)

  if binning.First.Lower != -100 {
    t.Error("test failed")
  }
  if binning.First.Upper != 1 {
    t.Error("test failed")
  }
  if binning.Largest.Lower != 120 {
    t.Error("test failed")
  }
  if binning.Largest.Upper != 300 {
    t.Error("test failed")
  }

  // delete first bin
  binning.Delete(&binning.Bins[0])
  binning.Update()

  if binning.Largest.Lower != -100 {
    t.Error("test failed")
  }
  if binning.Largest.Upper != 120 {
    t.Error("test failed")
  }
  if binning.Last.Lower != 350 {
    t.Error("test failed")
  }
  if binning.Last.Upper != 380 {
    t.Error("test failed")
  }

  // delete last bin
  binning.Delete(&binning.Bins[len(binning.Bins)-1])
  binning.Update()

  if binning.Last.Lower != 300 {
    t.Error("test failed")
  }
  if binning.Last.Upper != 380 {
    t.Error("test failed")
  }
}
