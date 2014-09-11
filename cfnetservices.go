package cfnetservices

/*
#cgo LDFLAGS: -framework CoreServices
#include <CFNetwork/CFNetServices.h>
#include <CFNetwork/CFHost.h>
#include <CoreFoundation/CFString.h>
#include <CoreFoundation/CFStream.h>
#include <stdio.h>
#include <stdlib.h>

extern void _CFNetServiceSetClientCallback(CFNetServiceRef, CFStreamError*, void*);
*/
import "C"
import (
	"fmt"
	"unsafe"
)

var (
	kCFStreamErrorDomainCustom              = int(C.kCFStreamErrorDomainCustom)
	kCFStreamErrorDomainPOSIX               = int(C.kCFStreamErrorDomainPOSIX)
	kCFStreamErrorDomainMacOSStatus         = int(C.kCFStreamErrorDomainMacOSStatus)
	kCFStreamErrorDomainNetDB               = int(C.kCFStreamErrorDomainNetDB)
	kCFStreamErrorDomainNetServices         = int(C.kCFStreamErrorDomainNetServices)
	kCFStreamErrorDomainSystemConfiguration = int(C.kCFStreamErrorDomainSystemConfiguration)
)

const (
	kCFNetServicesErrorUnknown     = int32(C.kCFNetServicesErrorUnknown)
	kCFNetServicesErrorCollision   = int32(C.kCFNetServicesErrorCollision)
	kCFNetServicesErrorNotFound    = int32(C.kCFNetServicesErrorNotFound)
	kCFNetServicesErrorInProgress  = int32(C.kCFNetServicesErrorInProgress)
	kCFNetServicesErrorBadArgument = int32(C.kCFNetServicesErrorBadArgument)
	kCFNetServicesErrorCancel      = int32(C.kCFNetServicesErrorCancel)
	kCFNetServicesErrorInvalid     = int32(C.kCFNetServicesErrorInvalid)
	kCFNetServicesErrorTimeout     = int32(C.kCFNetServicesErrorTimeout)
)

const (
	kCFNetServiceFlagNoAutoRename = C.kCFNetServiceFlagNoAutoRename
)

type CFNetService struct {
	ref C.CFNetServiceRef
	ctx *C.CFNetServiceClientContext
}

type CFStreamError struct {
	Domain int
	Code   int32
}

func (e *CFStreamError) Error() string {
	if e.Domain == kCFStreamErrorDomainNetServices {
		return fmt.Sprintf("%s: %d (%s)", getDomainString(e.Domain), e.Code, getNetServiceErrorString(e.Code))
	} else {
		return fmt.Sprintf("%s: %d", getDomainString(e.Domain), e.Code)
	}
}

func getDomainString(domain int) string {
	switch domain {
	case kCFStreamErrorDomainPOSIX:
		return "POSIX"
	case kCFStreamErrorDomainMacOSStatus:
		return "OSStatus"
	case kCFStreamErrorDomainNetDB:
		return "NetDB"
	case kCFStreamErrorDomainNetServices:
		return "NetServices"
	case kCFStreamErrorDomainSystemConfiguration:
		return "DomainSystemConfiguration"
	default:
		return "Custom"
	}
}

func getNetServiceErrorString(code int32) string {
	switch code {
	case kCFNetServicesErrorUnknown:
		return "unknown"
	case kCFNetServicesErrorCollision:
		return "collision"
	case kCFNetServicesErrorNotFound:
		return "not found"
	case kCFNetServicesErrorInProgress:
		return "in progress"
	case kCFNetServicesErrorBadArgument:
		return "bad argument"
	case kCFNetServicesErrorCancel:
		return "cancel"
	case kCFNetServicesErrorInvalid:
		return "invalid"
	case kCFNetServicesErrorTimeout:
		return "timeout"
	default:
		return "unknown"
	}
}

func buildCFStreamError(e *C.CFStreamError) *CFStreamError {
	return &CFStreamError{
		Domain: int(e.domain),
		Code:   int32(e.error),
	}
}

func NewCFString(s string) C.CFStringRef {
	s_ := C.CString(s)
	defer C.free(unsafe.Pointer(s_))
	retval := C.CFStringCreateWithBytes(
		C.CFAllocatorRef(nil), (*C.UInt8)(unsafe.Pointer(s_)), C.CFIndex(len(s)), C.kCFStringEncodingUTF8, C.Boolean(0),
	)
	return retval
}

func CFStringRelease(x C.CFStringRef) {
	C.CFRelease((C.CFTypeRef)(x))
}

func CFNetServiceCreate(domain string, serviceType string, name string, port int) *CFNetService {
	domain_ := NewCFString(domain)
	defer CFStringRelease(domain_)
	serviceType_ := NewCFString(serviceType)
	defer CFStringRelease(serviceType_)
	name_ := NewCFString(name)
	defer CFStringRelease(name_)
	return &CFNetService{
		ref: C.CFNetServiceCreate(nil, domain_, serviceType_, name_, C.SInt32(port)),
		ctx: nil,
	}
}

func CFNetServiceSetTXTData(cns *CFNetService, data []byte) bool {
	p := unsafe.Pointer(nil)
	if data != nil && len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	data_ := C.CFDataCreate(nil, (*C.UInt8)(p), C.CFIndex(len(data)))
	retval := C.CFNetServiceSetTXTData(cns.ref, data_)
	C.CFRelease((C.CFTypeRef)(data_))
	return retval != 0
}

//export _CFNetServiceSetClientCallback
func _CFNetServiceSetClientCallback(cns C.CFNetServiceRef, _ *C.CFStreamError, completionChan unsafe.Pointer) {
	*(*chan<- struct{})(completionChan) <- struct{}{}
}

func CFNetServiceRegisterWithOptions(cns *CFNetService, options int, completionChan chan<- struct{}) error {
	var e C.CFStreamError
	if completionChan != nil {
		ctx := &C.CFNetServiceClientContext{0, unsafe.Pointer(&completionChan), nil, nil, nil}
		cns.ctx = ctx
		C.CFNetServiceSetClient(cns.ref, (*[0]byte)(C._CFNetServiceSetClientCallback), ctx)
	}
	retval := C.CFNetServiceRegisterWithOptions(cns.ref, C.CFOptionFlags(options), &e)
	if retval != 0 {
		return nil
	} else {
		return buildCFStreamError(&e)
	}
}

func CFNetServiceCancel(cns *CFNetService) {
	C.CFNetServiceCancel(cns.ref)
}

func CFNetServiceRelease(cns *CFNetService) {
	C.CFRelease((C.CFTypeRef)(cns.ref))
}
