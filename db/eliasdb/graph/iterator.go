/*
 * EliasDB
 *
 * Copyright 2016 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package graph

import (
	"github.com/unix-world/smartgoext/db/eliasdb/graph/util"
	"github.com/unix-world/smartgoext/db/eliasdb/hash"
)

/*
NodeKeyIterator can be used to iterate node keys of a certain node kind.
*/
type NodeKeyIterator struct {
	gm        *Manager            // GraphManager which created the iterator
	it        *hash.HTreeIterator // Internal HTree iterator
	LastError error               // Last encountered error
}

/*
Next returns the next node key. Sets the LastError attribute if an error occurs.
*/
func (it *NodeKeyIterator) Next() string {

	// Take reader lock

	it.gm.mutex.RLock()
	defer it.gm.mutex.RUnlock()

	k, _ := it.it.Next()

	if it.it.LastError != nil {
		it.LastError = &util.GraphError{Type: util.ErrReading, Detail: it.it.LastError.Error()}
		return ""
	} else if len(k) == 0 {
		return ""
	}

	return string(k[len(PrefixNSAttrs):])
}

/*
HasNext returns if there is a next node key.
*/
func (it *NodeKeyIterator) HasNext() bool {
	return it.it.HasNext()
}

/*
Error returns the last encountered error.
*/
func (it *NodeKeyIterator) Error() error {
	return it.LastError
}

//-- unixman
/*
StatsCount returns a count of this index manager.
*/
func (it *NodeKeyIterator) StatsCount() uint64 {

	var cnt uint64 = 0

	if it.LastError != nil {
		return cnt
	}

	if(it.it == nil) {
		return cnt
	}
	for it.it.HasNext() {
		if it.it.LastError != nil {
			return cnt
		}
		cnt++
		it.it.Next() // IMPORTANT: advance 1 step, otherwise will fall into an infinite loop !
	}

	return cnt
}
//-- #end

//=====

//-- unixman

/*
EdgeKeyIterator can be used to iterate edge keys of a certain edge kind.
*/
type EdgeKeyIterator struct {
	gm        *Manager            // GraphManager which created the iterator
	it        *hash.HTreeIterator // Internal HTree iterator
	LastError error               // Last encountered error
}

/*
Next returns the next edge key. Sets the LastError attribute if an error occurs.
*/
func (it *EdgeKeyIterator) Next() string {

	// Take reader lock

	it.gm.mutex.RLock()
	defer it.gm.mutex.RUnlock()

	k, _ := it.it.Next()

	if it.it.LastError != nil {
		it.LastError = &util.GraphError{Type: util.ErrReading, Detail: it.it.LastError.Error()}
		return ""
	} else if len(k) == 0 {
		return ""
	}

	return string(k[len(PrefixNSAttrs):])
}

/*
HasNext returns if there is a next edge key.
*/
func (it *EdgeKeyIterator) HasNext() bool {
	return it.it.HasNext()
}

/*
Error returns the last encountered error.
*/
func (it *EdgeKeyIterator) Error() error {
	return it.LastError
}

/*
StatsCount returns a count of this index manager.
*/
func (it *EdgeKeyIterator) StatsCount() uint64 {

	var cnt uint64 = 0

	if it.LastError != nil {
		return cnt
	}

	if(it.it == nil) {
		return cnt
	}
	for it.it.HasNext() {
		if it.it.LastError != nil {
			return cnt
		}
		cnt++
		it.it.Next() // IMPORTANT: advance 1 step, otherwise will fall into an infinite loop !
	}

	return cnt
}

//-- end:unixman

// #end
