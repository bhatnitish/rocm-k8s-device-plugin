/**
 * Copyright 2018 Advanced Micro Devices, Inc.  All rights reserved.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
**/

package allocator

import (
	"math"
	"sort"
)

type bestEffortPolicy struct{}

func NewBestEffotPolicy() Policy {
	return &bestEffortPolicy{}
}

func getDevicesFromIds(total []*Device, ids []string) []*Device {
	var res []*Device
	for _, id := range ids {
		for _, dev := range total {
			if dev.Id == id {
				res = append(res, dev)
				break
			}
		}
	}
	return res
}

func (b *bestEffortPolicy) Allocate(availableIds, requiredIds []string, size int, devices []*Device) []string {
	if size <= 0 {
		return []string{}
	}

	if len(availableIds) < size {
		return []string{}
	}

	if len(requiredIds) > size {
		return []string{}
	}

	if len(requiredIds) > len(availableIds) {
		return []string{}
	}

	available := getDevicesFromIds(devices, availableIds)
	required := getDevicesFromIds(devices, requiredIds)

	p2pWeights := make(map[int]map[int]int)
	err := fetchAllPairWeights(available, p2pWeights, "")
	if err != nil {
		return []string{}
	}
	allSubsets := getAllDeviceSubsets(available, size, p2pWeights)

	var requiredNodeIds []int
	for i := 0; i < len(required); i++ {
		requiredNodeIds = append(requiredNodeIds, required[i].NodeId)
	}
	sort.Slice(requiredNodeIds, func(i, j int) bool {
		return i < j
	})

	bestScore := math.MaxInt32
	var candidate *DeviceSet
	for _, subset := range allSubsets {
		if !setContainsAll(subset.Ids, requiredNodeIds) {
			continue
		}
		if subset.TotalWeight < bestScore {
			candidate = subset
		}
	}
	if candidate == nil {
		return []string{}
	}
	var outset []string
	for _, id := range candidate.Ids {
		for _, d := range available {
			if d.NodeId == id {
				outset = append(outset, d.Id)
				break
			}
		}
	}
	return outset
}
