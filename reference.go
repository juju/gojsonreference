// Copyright 2013 sigu-399 ( https://github.com/sigu-399 )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// author  			sigu-399
// author-github 	https://github.com/sigu-399
// author-mail		sigu.399@gmail.com
//
// repository-name	gojsonreference
// repository-desc	An implementation of JSON Reference - Go language
//
// description		Main and unique file.
//
// created      	26-02-2013

package gojsonreference

import (
	"errors"
	"net/url"
	"strings"

	"github.com/binary132/gojsonpointer"
)

const (
	const_fragment_char = `#`
)

func NewJsonReference(jsonReferenceString string) (JsonReference, error) {
	var r JsonReference
	err := r.parse(jsonReferenceString)
	return r, err

}

type JsonReference struct {
	referenceUrl     *url.URL
	referencePointer gojsonpointer.JsonPointer

	HasFragmentOnly bool
}

func (r *JsonReference) GetUrl() *url.URL {
	return r.referenceUrl
}

func (r *JsonReference) GetPointer() *gojsonpointer.JsonPointer {
	return &r.referencePointer
}

func (r *JsonReference) String() string {

	if r.HasFragmentOnly {
		return const_fragment_char + r.referencePointer.String()
	}

	if r.referenceUrl != nil {
		return r.referenceUrl.String()
	}

	println("Never get here!")
	return r.referencePointer.String()
}

// "Constructor", parses the given string JSON reference
func (r *JsonReference) parse(jsonReferenceString string) error {

	var err error

	// fragment only
	if strings.HasPrefix(jsonReferenceString, const_fragment_char) {
		r.referencePointer, err = gojsonpointer.NewJsonPointer(jsonReferenceString[1:])
		if err != nil {
			return err
		}

		r.referenceUrl, err = url.Parse(jsonReferenceString)
		if err != nil {
			return err
		}

		r.HasFragmentOnly = true
	} else {
		r.referenceUrl, err = url.Parse(jsonReferenceString)
		if err != nil {
			return err
		}

		r.referencePointer, err = gojsonpointer.NewJsonPointer(r.referenceUrl.Fragment)
		if err != nil {
			return err
		}
	}

	return nil
}

// Creates a new reference from a parent and a child
// If the child cannot inherit from the parent, an error is returned
func (r *JsonReference) Inherits(child JsonReference) (*JsonReference, error) {

	if !child.HasFragmentOnly {
		if child.GetUrl().Scheme != r.GetUrl().Scheme {
			return nil, errors.New("Scheme of child " + child.String() +
				"incompatible with scheme of parent " + r.String())
		}
	}

	switch r.GetUrl().Scheme {
	case "http":
		return r.inheritsImplHttp(child)
	case "file":
		return r.inheritsImplFile(child)
	}

	return nil, errors.New("Scheme type " + r.GetUrl().Scheme + " not handled.")
}

func (r *JsonReference) inheritsImplFile(child JsonReference) (*JsonReference, error) {

	if !r.GetUrl().IsAbs() {
		return nil, errors.New("Parent reference must be absolute URL.")
	}

	childReference := child.String()

	if child.HasFragmentOnly {
		childReference = r.GetUrl().Scheme + "://" + r.GetUrl().Path + child.String()
	}

	inheritedReference, err := NewJsonReference(childReference)
	if err != nil {
		return nil, err
	}

	return &inheritedReference, nil
}

func (r *JsonReference) inheritsImplHttp(child JsonReference) (*JsonReference, error) {

	if !r.referenceUrl.IsAbs() {
		return nil, errors.New("Parent reference must be absolute URL.")
	}

	if r.referenceUrl.IsAbs() && child.referenceUrl.IsAbs() {
		println("Got inside child = " + child.String())
		if r.referenceUrl.Scheme != child.referenceUrl.Scheme {
			return nil, errors.New("References have different schemes")
		}
		if r.referenceUrl.Host != child.referenceUrl.Host {
			return nil, errors.New("References have different hosts")
		}
	}

	inheritedReference, err := NewJsonReference(r.String())
	if err != nil {
		return nil, err
	}

	if child.HasFragmentOnly {
		inheritedReference.referenceUrl.Fragment = child.referencePointer.String()
		inheritedReference.referencePointer = child.referencePointer
	}
	if !child.referenceUrl.IsAbs() {
		inheritedReference.referenceUrl.Path = child.referenceUrl.Path
	} else {
		inheritedReference.referenceUrl.Fragment = child.referenceUrl.Fragment
		inheritedReference.referenceUrl.Path = child.referenceUrl.Path
	}
	return &inheritedReference, nil
}
