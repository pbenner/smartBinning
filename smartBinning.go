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

import   "fmt"
import   "math"
import   "sort"

/* -------------------------------------------------------------------------- */

type Bin struct {
  Y        float64
  Lower    float64
  Upper    float64
  Next    *Bin
  Prev    *Bin
  Smaller *Bin
  Larger  *Bin
  Deleted  bool
}

func (bin Bin) Size() float64 {
  return bin.Upper - bin.Lower
}

func (bin Bin) String() string {
  return fmt.Sprintf("[%f, %f):%v", bin.Lower, bin.Upper, bin.Y)
}

/* -------------------------------------------------------------------------- */

type binList []Bin

func (bins binList) Len() int {
  return len(bins)
}

func (bins binList) Less(i, j int) bool {
  return bins[i].Lower < bins[j].Lower
}

func (bins binList) Swap(i, j int) {
  bins[i], bins[j] = bins[j], bins[i]
}

/* -------------------------------------------------------------------------- */

type binListSorted struct {
  bins []*Bin
  less func(Bin, Bin) bool
}

func (obj binListSorted) Len() int {
  return len(obj.bins)
}

func (obj binListSorted) Less(i, j int) bool {
  return obj.less(*obj.bins[i], *obj.bins[j])
}

func (obj binListSorted) Swap(i, j int) {
  obj.bins[i], obj.bins[j] = obj.bins[j], obj.bins[i]
}

/* -------------------------------------------------------------------------- */

func BinLessSize(a, b Bin) bool {
  return a.Size() < b.Size()
}

func BinLessY(a, b Bin) bool {
  return a.Y < b.Y
}

func BinSum(a, b Bin) float64 {
  return a.Y + b.Y
}

func BinLogSum(a, b Bin) float64 {
  x, y :=  a.Y, b.Y
  if x > y {
    // swap
    x, y = x, y
  }
  if math.IsInf(x, -1) {
    return y
  }
  return y + math.Log1p(math.Exp(x-y))
}

/* -------------------------------------------------------------------------- */

type Binning struct {
  Bins      binList
  Sum       func(Bin, Bin) float64
  Less      func(Bin, Bin) bool
  First    *Bin
  Last     *Bin
  Smallest *Bin
  Largest  *Bin
  Verbose   bool
}

func New(x, y []float64, sum func(Bin, Bin) float64, less func(Bin, Bin) bool) (*Binning, error) {
  n := len(x)-1

  if n < 2 {
    return nil, fmt.Errorf("length of x must be greater than two")
  }
  binning := Binning{}
  binning.Bins = make(binList, n)
  binning.Sum  = sum
  binning.Less = less
  bins := make([]*Bin, n)

  // set lower boundaries
  for i := 0; i < n; i++ {
    binning.Bins[i].Lower = x[i]
  }
  // set y
  switch len(y) {
  case 0:
  case 1:
    for i := 0; i < n; i++ {
      binning.Bins[i].Y = y[0]
    }
  default:
    if len(y) != n {
      return nil, fmt.Errorf("y vector has invalid length")
    }
    for i := 0; i < n; i++ {
      binning.Bins[i].Y = y[i]
    }
  }
  // get bins in the right order
  sort.Sort(binning.Bins)
  // set upper boundaries
  for i := 0; i < n-1; i++ {
    binning.Bins[i].Upper = binning.Bins[i+1].Lower
  }
  binning.Bins[n-1].Upper = x[n]
  // create linked lists
  for i := 0; i < n-1; i++ {
    binning.Bins[i].Next = &binning.Bins[i+1]
  }
  for i := 1; i < n; i++ {
    binning.Bins[i].Prev = &binning.Bins[i-1]
  }
  binning.First = &binning.Bins[0]
  binning.Last  = &binning.Bins[n-1]
  // create a binList and sort the elements
  for i := 0; i < n; i++ {
    bins[i] = &binning.Bins[i]
  }
  sort.Sort(binListSorted{bins, binning.Less})

  for i := 0; i < len(bins)-1; i++ {
    bins[i].Larger = bins[i+1]
  }
  for i := 1; i < len(bins); i++ {
    bins[i].Smaller = bins[i-1]
  }
  binning.Smallest = bins[0]
  binning.Largest  = bins[n-1]

  return &binning, nil
}

func (binning *Binning) deleteBin(bin *Bin) *Bin {
  // delete from linked list
  if bin.Prev != nil && bin.Next != nil {
    bin.Prev.Next = bin.Next
    bin.Next.Prev = bin.Prev
  } else {
    if bin.Prev != nil {
      // deleting last bin
      bin.Prev.Next = nil
      binning.Last = bin.Prev
    }
    if bin.Next != nil {
      // deleting first bin
      bin.Next.Prev = nil
      binning.First = bin.Next
    }
  }
  // delete from sorted linked list
  binning.deleteBinSorted(bin)
  // mark bin as deleted
  bin.Deleted = true
  // merge bin data
  if bin.Prev == nil {
    // there is no bin to the left, merge
    // with bin on the right
    bin.Next.Y     = binning.Sum(*bin.Next, *bin)
    bin.Next.Lower = bin.Lower
    bin = bin.Next
  } else
  if bin.Next == nil {
    // there is no bin to the right, merge
    // with bin on the left
    bin.Prev.Y     = binning.Sum(*bin.Prev, *bin)
    bin.Prev.Upper = bin.Upper
    bin = bin.Prev
  } else {
    // merge bin with smaller bin around
    if binning.Less(*bin.Prev, *bin.Next) {
      // merge with bin to the left
      bin.Prev.Y     = binning.Sum(*bin.Prev, *bin)
      bin.Prev.Upper = bin.Upper
      bin = bin.Prev
    } else {
      // merge with bin to the right
      bin.Next.Y     = binning.Sum(*bin.Next, *bin)
      bin.Next.Lower = bin.Lower
      bin = bin.Next
    }
  }
  return bin
}

func (binning *Binning) deleteBinSorted(bin *Bin) {
  if bin.Smaller != nil && bin.Larger != nil {
    bin.Smaller.Larger = bin.Larger
    bin.Larger.Smaller = bin.Smaller
  } else {
    if bin.Smaller != nil {
      // deleting largest bin
      bin.Smaller.Larger = nil
      binning.Largest = bin.Smaller
    }
    if bin.Larger != nil {
      // deleting smallest bin
      bin.Larger.Smaller = nil
      binning.Smallest = bin.Larger
    }
  }
}

func (binning *Binning) insertBinSortedBefore(bin, at *Bin) {
  if at.Smaller != nil {
    at.Smaller.Larger = bin
  }
  bin.Smaller = at.Smaller
  bin.Larger  = at
  at.Smaller  = bin
}

func (binning *Binning) insertBinSortedAfter(bin, at *Bin) {
  if at.Larger != nil {
    at.Larger.Smaller = bin
  }
  bin.Smaller = at
  bin.Larger  = at.Larger
  at.Larger   = bin
}

func (binning *Binning) Delete(bin *Bin) {
  if bin.Prev == nil && bin.Next == nil {
    return
  }
  // delete bin from linked list
  bin = binning.deleteBin(bin)
  // update bin size
  if bin.Larger != nil && binning.Less(*bin.Larger, *bin) {
    // save next largest bin as current position
    at := bin.Larger
    // delete bin from sorted list
    binning.deleteBinSorted(bin)
    // find new position for the bin
    for at.Larger != nil && binning.Less(*at, *bin) {
      at = at.Larger
    }
    if binning.Less(*bin, *at) {
      binning.insertBinSortedBefore(bin, at)
    } else {
      binning.insertBinSortedAfter(bin, at)
    }
  }
}

func (binning *Binning) Update() error {
  // get new values
  x := []float64{}
  y := []float64{}
  for t := binning.First; t != nil; t = t.Next {
    if t.Deleted {
      // this shouldn't happen
      panic("internal error")
    }
    x = append(x, t.Lower)
    y = append(y, t.Y)
  }
  x = append(x, binning.Bins[len(binning.Bins)-1].Upper)

  if tmp, err := New(x, y, binning.Sum, binning.Less); err != nil {
    return err
  } else {
    *binning = *tmp
  }
  return nil
}

func (binning *Binning) FilterBins(n int) error {
  if len(binning.Bins) == 0 || len(binning.Bins) < n {
    return nil
  }
  m := len(binning.Bins) - n
  for i := 0; i < m; i++ {
    binning.Delete(binning.Smallest)
  }
  return binning.Update()
}
