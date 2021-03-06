// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include "version.h"
*/
import "C"
import (
	"unsafe"
)

type defInt8 struct {
	rset       *Rset
	ocidef     *C.OCIDefine
	ociNumber  [1]C.OCINumber
	isNullable bool
	nullp
}

func (def *defInt8) define(position int, isNullable bool, rset *Rset) error {
	def.rset = rset
	def.isNullable = isNullable
	r := C.OCIDEFINEBYPOS(
		def.rset.ocistmt,                    //OCIStmt     *stmtp,
		&def.ocidef,                         //OCIDefine   **defnpp,
		def.rset.stmt.ses.srv.env.ocierr,    //OCIError    *errhp,
		C.ub4(position),                     //ub4         position,
		unsafe.Pointer(&def.ociNumber[0]),   //void        *valuep,
		C.LENGTH_TYPE(C.sizeof_OCINumber),   //sb8         value_sz,
		C.SQLT_VNU,                          //ub2         dty,
		unsafe.Pointer(def.nullp.Pointer()), //void        *indp,
		nil,           //ub2         *rlenp,
		nil,           //ub2         *rcodep,
		C.OCI_DEFAULT) //ub4         mode );
	if r == C.OCI_ERROR {
		return def.rset.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (def *defInt8) value() (value interface{}, err error) {
	if def.isNullable {
		oraInt8Value := Int8{IsNull: def.nullp.IsNull()}
		if !oraInt8Value.IsNull {
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr,    //OCIError              *err,
				&def.ociNumber[0],                   //const OCINumber       *number,
				C.uword(1),                          //uword                 rsl_length,
				C.OCI_NUMBER_SIGNED,                 //uword                 rsl_flag,
				unsafe.Pointer(&oraInt8Value.Value)) //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
		}
		value = oraInt8Value
	} else {
		var int8Value int8
		if !def.nullp.IsNull() {
			r := C.OCINumberToInt(
				def.rset.stmt.ses.srv.env.ocierr, //OCIError              *err,
				&def.ociNumber[0],                //const OCINumber       *number,
				C.uword(1),                       //uword                 rsl_length,
				C.OCI_NUMBER_SIGNED,              //uword                 rsl_flag,
				unsafe.Pointer(&int8Value))       //void                  *rsl );
			if r == C.OCI_ERROR {
				err = def.rset.stmt.ses.srv.env.ociError()
			}
		}
		value = int8Value
	}
	return value, err
}

func (def *defInt8) alloc() error {
	return nil
}

func (def *defInt8) free() {
}

func (def *defInt8) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	rset := def.rset
	def.rset = nil
	def.ocidef = nil
	def.nullp.Free()
	rset.putDef(defIdxInt8, def)
	return nil
}
