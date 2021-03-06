// Copyright 2014 Rana Ian. All rights reserved.
// Use of this source code is governed by The MIT License
// found in the accompanying LICENSE file.

package ora

/*
#include <oci.h>
#include <stdlib.h>
#include "version.h"
*/
import "C"
import (
	"bytes"
	"time"
	"unsafe"
)

type bndTimePtr struct {
	stmt    *Stmt
	ocibnd  *C.OCIBind
	value   *time.Time
	cZone   *C.char
	zoneBuf bytes.Buffer
	dateTimep
	nullp
}

func (bnd *bndTimePtr) bind(value *time.Time, position int, stmt *Stmt) error {
	bnd.stmt = stmt
	bnd.value = value
	r := C.OCIDescriptorAlloc(
		unsafe.Pointer(bnd.stmt.ses.srv.env.ocienv),                //CONST dvoid   *parenth,
		(*unsafe.Pointer)(unsafe.Pointer(bnd.dateTimep.Pointer())), //dvoid         **descpp,
		C.OCI_DTYPE_TIMESTAMP_TZ,                                   //ub4           type,
		0,   //size_t        xtramem_sz,
		nil) //dvoid         **usrmempp);
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	} else if r == C.OCI_INVALID_HANDLE {
		return errNew("unable to allocate oci timestamp handle during bind")
	}
	bnd.nullp.Set(value == nil)
	if value != nil {
		zone := zoneOffset(*value, &bnd.zoneBuf)
		bnd.cZone = C.CString(zone)
		r = C.OCIDateTimeConstruct(
			unsafe.Pointer(bnd.stmt.ses.srv.env.ocienv), //dvoid         *hndl,
			bnd.stmt.ses.srv.env.ocierr,                 //OCIError      *err,
			bnd.dateTimep.Value(),                       //OCIDateTime   *datetime,
			C.sb2(value.Year()),                         //sb2           year,
			C.ub1(int32(value.Month())),                 //ub1           month,
			C.ub1(value.Day()),                          //ub1           day,
			C.ub1(value.Hour()),                         //ub1           hour,
			C.ub1(value.Minute()),                       //ub1           min,
			C.ub1(value.Second()),                       //ub1           sec,
			C.ub4(value.Nanosecond()),                   //ub4           fsec,
			(*C.OraText)(unsafe.Pointer(bnd.cZone)),     //OraText       *timezone,
			C.size_t(len(zone)))                         //size_t        timezone_length );
		if r == C.OCI_ERROR {
			return bnd.stmt.ses.srv.env.ociError()
		}
	}
	r = C.OCIBINDBYPOS(
		bnd.stmt.ocistmt, //OCIStmt      *stmtp,
		&bnd.ocibnd,
		bnd.stmt.ses.srv.env.ocierr,             //OCIError     *errhp,
		C.ub4(position),                         //ub4          position,
		unsafe.Pointer(bnd.dateTimep.Pointer()), //void         *valuep,
		C.LENGTH_TYPE(bnd.dateTimep.Size()),     //sb8          value_sz,
		C.SQLT_TIMESTAMP_TZ,                     //ub2          dty,
		unsafe.Pointer(bnd.nullp.Pointer()),     //void         *indp,
		nil,           //ub2          *alenp,
		nil,           //ub2          *rcodep,
		0,             //ub4          maxarr_len,
		nil,           //ub4          *curelep,
		C.OCI_DEFAULT) //ub4          mode );
	if r == C.OCI_ERROR {
		return bnd.stmt.ses.srv.env.ociError()
	}
	return nil
}

func (bnd *bndTimePtr) setPtr() (err error) {
	if bnd.value == nil { // cannot set on a nil pointer
		return nil
	}
	if bnd.nullp.IsNull() {
		*bnd.value = time.Time{} // zero time
		return nil
	}
	*bnd.value, err = getTime(bnd.stmt.ses.srv.env, bnd.dateTimep.Value())
	return err
}

func (bnd *bndTimePtr) close() (err error) {
	defer func() {
		if value := recover(); value != nil {
			err = errR(value)
		}
	}()

	if bnd.cZone != nil {
		C.free(unsafe.Pointer(bnd.cZone))
		bnd.cZone = nil
		C.OCIDescriptorFree(
			unsafe.Pointer(bnd.dateTimep.Value()), //void     *descp,
			C.OCI_DTYPE_TIMESTAMP_TZ)              //ub4      type );
	}
	stmt := bnd.stmt
	bnd.stmt = nil
	bnd.ocibnd = nil
	bnd.value = nil
	bnd.zoneBuf.Reset()
	bnd.dateTimep.Free()
	bnd.nullp.Free()
	stmt.putBnd(bndIdxTimePtr, bnd)
	return nil
}
