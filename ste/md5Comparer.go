// Copyright © 2017 Microsoft <wastore@microsoft.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package ste

import (
	"bytes"
	"errors"
	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-azcopy/common"
)

type transferSpecificLogger interface {
	LogAtLevelForCurrentTransfer(level pipeline.LogLevel, msg string)
}

type md5Comparer struct {
	expected         []byte
	actualAsSaved    []byte
	validationOption common.HashValidationOption
	logger           transferSpecificLogger
}

// TODO: let's add an aka.ms link to the message,  that gives more info
var errMd5Mismatch = errors.New("the MD5 hash of the data, as we received it, did not match the expected value, as found in the Blob/File Service. " +
	"This means that either there is a data integrity error OR another tool has failed to keep the stored hash up to date")

// TODO: let's add an aka.ms link to the message, that gives more info
const noMD5Stored = "no MD5 was stored in the Blob/File service against this file. So the downloaded data cannot be MD5-validated."

var errExpectedMd5Missing = errors.New(noMD5Stored + " This application is currently configured to treat missing MD5 hashes as errors")

var errActualMd5NotComputed = errors.New("no MDB was computed within this application. This indicates a logic error in this application")

// Check compares the two MD5s, and returns any error if applicable
// Any informational logging will be done within Check, so all the caller needs to do
// is respond to non-nil errors
func (c *md5Comparer) Check() error {

	if c.validationOption == common.EHashValidationOption.NoCheck() {
		return nil
	}

	if c.actualAsSaved == nil || len(c.actualAsSaved) == 0 {
		return errActualMd5NotComputed // Should never happen, so there's no way to opt out of this error being returned if it DOES happen
	}

	// missing
	if c.expected == nil || len(c.expected) == 0 {
		switch c.validationOption {
		case common.EHashValidationOption.FailIfDifferentOrMissing():
			return errExpectedMd5Missing
		case common.EHashValidationOption.FailIfDifferent(),
			common.EHashValidationOption.LogOnly():
			c.logAsMissing()
			return nil
		default:
			panic("unexpected hash validation type")
		}
	}

	// different
	match := bytes.Equal(c.expected, c.actualAsSaved)
	if !match {
		switch c.validationOption {
		case common.EHashValidationOption.FailIfDifferentOrMissing(),
			common.EHashValidationOption.FailIfDifferent():
			return errMd5Mismatch
		case common.EHashValidationOption.LogOnly():
			c.logAsDifferent()
			return nil
		default:
			panic("unexpected hash validation type")
		}
	}

	return nil
}

func (c *md5Comparer) logAsMissing() {
	c.logger.LogAtLevelForCurrentTransfer(pipeline.LogWarning, noMD5Stored)
}

func (c *md5Comparer) logAsDifferent() {
	c.logger.LogAtLevelForCurrentTransfer(pipeline.LogWarning, errMd5Mismatch.Error())
}
