// Code generated by github.com/ngicks/go-codegen/codegen/generator/cloner/internal/liststdtypes DO NOT EDIT.
package cloner

import (
	"github.com/ngicks/go-codegen/codegen/imports"
)

var stdCloneByAssign = map[imports.TargetType]struct{}{
	{ImportPath: "crypto/ed25519", Name: "Options"}:                        {},
	{ImportPath: "crypto/rc4", Name: "Cipher"}:                             {},
	{ImportPath: "crypto/rsa", Name: "PKCS1v15DecryptOptions"}:             {},
	{ImportPath: "crypto/rsa", Name: "PSSOptions"}:                         {},
	{ImportPath: "crypto/x509", Name: "ConstraintViolationError"}:          {},
	{ImportPath: "crypto/x509", Name: "UnhandledCriticalExtension"}:        {},
	{ImportPath: "database/sql", Name: "DBStats"}:                          {},
	{ImportPath: "database/sql", Name: "NullBool"}:                         {},
	{ImportPath: "database/sql", Name: "NullByte"}:                         {},
	{ImportPath: "database/sql", Name: "NullFloat64"}:                      {},
	{ImportPath: "database/sql", Name: "NullInt16"}:                        {},
	{ImportPath: "database/sql", Name: "NullInt32"}:                        {},
	{ImportPath: "database/sql", Name: "NullInt64"}:                        {},
	{ImportPath: "database/sql", Name: "NullString"}:                       {},
	{ImportPath: "database/sql", Name: "TxOptions"}:                        {},
	{ImportPath: "database/sql/driver", Name: "Bool"}:                      {},
	{ImportPath: "database/sql/driver", Name: "DefaultParameterConverter"}: {},
	{ImportPath: "database/sql/driver", Name: "Int32"}:                     {},
	{ImportPath: "database/sql/driver", Name: "ResultNoRows"}:              {},
	{ImportPath: "database/sql/driver", Name: "String"}:                    {},
	{ImportPath: "database/sql/driver", Name: "TxOptions"}:                 {},
	{ImportPath: "debug/dwarf", Name: "AddrType"}:                          {},
	{ImportPath: "debug/dwarf", Name: "BasicType"}:                         {},
	{ImportPath: "debug/dwarf", Name: "BoolType"}:                          {},
	{ImportPath: "debug/dwarf", Name: "CharType"}:                          {},
	{ImportPath: "debug/dwarf", Name: "CommonType"}:                        {},
	{ImportPath: "debug/dwarf", Name: "ComplexType"}:                       {},
	{ImportPath: "debug/dwarf", Name: "DecodeError"}:                       {},
	{ImportPath: "debug/dwarf", Name: "DotDotDotType"}:                     {},
	{ImportPath: "debug/dwarf", Name: "EnumValue"}:                         {},
	{ImportPath: "debug/dwarf", Name: "FloatType"}:                         {},
	{ImportPath: "debug/dwarf", Name: "IntType"}:                           {},
	{ImportPath: "debug/dwarf", Name: "LineFile"}:                          {},
	{ImportPath: "debug/dwarf", Name: "UcharType"}:                         {},
	{ImportPath: "debug/dwarf", Name: "UintType"}:                          {},
	{ImportPath: "debug/dwarf", Name: "UnspecifiedType"}:                   {},
	{ImportPath: "debug/dwarf", Name: "UnsupportedType"}:                   {},
	{ImportPath: "debug/dwarf", Name: "VoidType"}:                          {},
	{ImportPath: "debug/elf", Name: "Chdr32"}:                              {},
	{ImportPath: "debug/elf", Name: "Chdr64"}:                              {},
	{ImportPath: "debug/elf", Name: "Dyn32"}:                               {},
	{ImportPath: "debug/elf", Name: "Dyn64"}:                               {},
	{ImportPath: "debug/elf", Name: "Header32"}:                            {},
	{ImportPath: "debug/elf", Name: "Header64"}:                            {},
	{ImportPath: "debug/elf", Name: "ImportedSymbol"}:                      {},
	{ImportPath: "debug/elf", Name: "Prog32"}:                              {},
	{ImportPath: "debug/elf", Name: "Prog64"}:                              {},
	{ImportPath: "debug/elf", Name: "ProgHeader"}:                          {},
	{ImportPath: "debug/elf", Name: "Rel32"}:                               {},
	{ImportPath: "debug/elf", Name: "Rel64"}:                               {},
	{ImportPath: "debug/elf", Name: "Rela32"}:                              {},
	{ImportPath: "debug/elf", Name: "Rela64"}:                              {},
	{ImportPath: "debug/elf", Name: "Section32"}:                           {},
	{ImportPath: "debug/elf", Name: "Section64"}:                           {},
	{ImportPath: "debug/elf", Name: "SectionHeader"}:                       {},
	{ImportPath: "debug/elf", Name: "Sym32"}:                               {},
	{ImportPath: "debug/elf", Name: "Sym64"}:                               {},
	{ImportPath: "debug/elf", Name: "Symbol"}:                              {},
	{ImportPath: "debug/gosym", Name: "UnknownLineError"}:                  {},
	{ImportPath: "debug/macho", Name: "DylibCmd"}:                          {},
	{ImportPath: "debug/macho", Name: "DysymtabCmd"}:                       {},
	{ImportPath: "debug/macho", Name: "FatArchHeader"}:                     {},
	{ImportPath: "debug/macho", Name: "FileHeader"}:                        {},
	{ImportPath: "debug/macho", Name: "Nlist32"}:                           {},
	{ImportPath: "debug/macho", Name: "Nlist64"}:                           {},
	{ImportPath: "debug/macho", Name: "Regs386"}:                           {},
	{ImportPath: "debug/macho", Name: "RegsAMD64"}:                         {},
	{ImportPath: "debug/macho", Name: "Reloc"}:                             {},
	{ImportPath: "debug/macho", Name: "RpathCmd"}:                          {},
	{ImportPath: "debug/macho", Name: "Section32"}:                         {},
	{ImportPath: "debug/macho", Name: "Section64"}:                         {},
	{ImportPath: "debug/macho", Name: "SectionHeader"}:                     {},
	{ImportPath: "debug/macho", Name: "Segment32"}:                         {},
	{ImportPath: "debug/macho", Name: "Segment64"}:                         {},
	{ImportPath: "debug/macho", Name: "SegmentHeader"}:                     {},
	{ImportPath: "debug/macho", Name: "Symbol"}:                            {},
	{ImportPath: "debug/macho", Name: "SymtabCmd"}:                         {},
	{ImportPath: "debug/pe", Name: "COFFSymbol"}:                           {},
	{ImportPath: "debug/pe", Name: "COFFSymbolAuxFormat5"}:                 {},
	{ImportPath: "debug/pe", Name: "DataDirectory"}:                        {},
	{ImportPath: "debug/pe", Name: "FileHeader"}:                           {},
	{ImportPath: "debug/pe", Name: "FormatError"}:                          {},
	{ImportPath: "debug/pe", Name: "ImportDirectory"}:                      {},
	{ImportPath: "debug/pe", Name: "OptionalHeader32"}:                     {},
	{ImportPath: "debug/pe", Name: "OptionalHeader64"}:                     {},
	{ImportPath: "debug/pe", Name: "Reloc"}:                                {},
	{ImportPath: "debug/pe", Name: "SectionHeader"}:                        {},
	{ImportPath: "debug/pe", Name: "SectionHeader32"}:                      {},
	{ImportPath: "debug/pe", Name: "Symbol"}:                               {},
	{ImportPath: "debug/plan9obj", Name: "FileHeader"}:                     {},
	{ImportPath: "debug/plan9obj", Name: "SectionHeader"}:                  {},
	{ImportPath: "debug/plan9obj", Name: "Sym"}:                            {},
	{ImportPath: "encoding/asn1", Name: "StructuralError"}:                 {},
	{ImportPath: "encoding/asn1", Name: "SyntaxError"}:                     {},
	{ImportPath: "encoding/base32", Name: "Encoding"}:                      {},
	{ImportPath: "encoding/base64", Name: "Encoding"}:                      {},
	{ImportPath: "encoding/binary", Name: "BigEndian"}:                     {},
	{ImportPath: "encoding/binary", Name: "LittleEndian"}:                  {},
	{ImportPath: "encoding/binary", Name: "NativeEndian"}:                  {},
	{ImportPath: "encoding/gob", Name: "CommonType"}:                       {},
	{ImportPath: "encoding/json", Name: "InvalidUTF8Error"}:                {},
	{ImportPath: "encoding/json", Name: "SyntaxError"}:                     {},
	{ImportPath: "encoding/xml", Name: "Attr"}:                             {},
	{ImportPath: "encoding/xml", Name: "EndElement"}:                       {},
	{ImportPath: "encoding/xml", Name: "Name"}:                             {},
	{ImportPath: "encoding/xml", Name: "SyntaxError"}:                      {},
	{ImportPath: "go/ast", Name: "BadDecl"}:                                {},
	{ImportPath: "go/ast", Name: "BadExpr"}:                                {},
	{ImportPath: "go/ast", Name: "BadStmt"}:                                {},
	{ImportPath: "go/ast", Name: "BasicLit"}:                               {},
	{ImportPath: "go/ast", Name: "Comment"}:                                {},
	{ImportPath: "go/ast", Name: "EmptyStmt"}:                              {},
	{ImportPath: "go/build", Name: "Directive"}:                            {},
	{ImportPath: "go/build", Name: "NoGoError"}:                            {},
	{ImportPath: "go/build/constraint", Name: "SyntaxError"}:               {},
	{ImportPath: "go/build/constraint", Name: "TagExpr"}:                   {},
	{ImportPath: "go/doc", Name: "Note"}:                                   {},
	{ImportPath: "go/doc/comment", Name: "Code"}:                           {},
	{ImportPath: "go/doc/comment", Name: "LinkDef"}:                        {},
	{ImportPath: "go/printer", Name: "Config"}:                             {},
	{ImportPath: "go/scanner", Name: "Error"}:                              {},
	{ImportPath: "go/token", Name: "Position"}:                             {},
	{ImportPath: "go/types", Name: "Basic"}:                                {},
	{ImportPath: "go/types", Name: "StdSizes"}:                             {},
	{ImportPath: "hash/crc32", Name: "Table"}:                              {},
	{ImportPath: "hash/crc64", Name: "Table"}:                              {},
	{ImportPath: "hash/maphash", Name: "Seed"}:                             {},
	{ImportPath: "image", Name: "Point"}:                                   {},
	{ImportPath: "image", Name: "Rectangle"}:                               {},
	{ImportPath: "image", Name: "ZP"}:                                      {},
	{ImportPath: "image", Name: "ZR"}:                                      {},
	{ImportPath: "image/color", Name: "Alpha"}:                             {},
	{ImportPath: "image/color", Name: "Alpha16"}:                           {},
	{ImportPath: "image/color", Name: "Black"}:                             {},
	{ImportPath: "image/color", Name: "CMYK"}:                              {},
	{ImportPath: "image/color", Name: "Gray"}:                              {},
	{ImportPath: "image/color", Name: "Gray16"}:                            {},
	{ImportPath: "image/color", Name: "NRGBA"}:                             {},
	{ImportPath: "image/color", Name: "NRGBA64"}:                           {},
	{ImportPath: "image/color", Name: "NYCbCrA"}:                           {},
	{ImportPath: "image/color", Name: "Opaque"}:                            {},
	{ImportPath: "image/color", Name: "RGBA"}:                              {},
	{ImportPath: "image/color", Name: "RGBA64"}:                            {},
	{ImportPath: "image/color", Name: "Transparent"}:                       {},
	{ImportPath: "image/color", Name: "White"}:                             {},
	{ImportPath: "image/color", Name: "YCbCr"}:                             {},
	{ImportPath: "image/jpeg", Name: "Options"}:                            {},
	{ImportPath: "log/slog", Name: "Source"}:                               {},
	{ImportPath: "math/big", Name: "ErrNaN"}:                               {},
	{ImportPath: "math/rand/v2", Name: "ChaCha8"}:                          {},
	{ImportPath: "math/rand/v2", Name: "PCG"}:                              {},
	{ImportPath: "net", Name: "AddrError"}:                                 {},
	{ImportPath: "net", Name: "KeepAliveConfig"}:                           {},
	{ImportPath: "net", Name: "MX"}:                                        {},
	{ImportPath: "net", Name: "NS"}:                                        {},
	{ImportPath: "net", Name: "ParseError"}:                                {},
	{ImportPath: "net", Name: "SRV"}:                                       {},
	{ImportPath: "net", Name: "UnixAddr"}:                                  {},
	{ImportPath: "net/http", Name: "MaxBytesError"}:                        {},
	{ImportPath: "net/http", Name: "NoBody"}:                               {},
	{ImportPath: "net/http", Name: "ProtocolError"}:                        {},
	{ImportPath: "net/http/httptrace", Name: "DNSStartInfo"}:               {},
	{ImportPath: "net/mail", Name: "Address"}:                              {},
	{ImportPath: "net/textproto", Name: "Error"}:                           {},
	{ImportPath: "net/url", Name: "Userinfo"}:                              {},
	{ImportPath: "os/user", Name: "Group"}:                                 {},
	{ImportPath: "os/user", Name: "User"}:                                  {},
	{ImportPath: "reflect", Name: "SliceHeader"}:                           {},
	{ImportPath: "reflect", Name: "StringHeader"}:                          {},
	{ImportPath: "reflect", Name: "ValueError"}:                            {},
	{ImportPath: "regexp/syntax", Name: "Error"}:                           {},
	{ImportPath: "runtime", Name: "BlockProfileRecord"}:                    {},
	{ImportPath: "runtime", Name: "Func"}:                                  {},
	{ImportPath: "runtime", Name: "MemProfileRecord"}:                      {},
	{ImportPath: "runtime", Name: "MemStats"}:                              {},
	{ImportPath: "runtime", Name: "StackRecord"}:                           {},
	{ImportPath: "runtime/cgo", Name: "Incomplete"}:                        {},
	{ImportPath: "runtime/debug", Name: "BuildSetting"}:                    {},
	{ImportPath: "runtime/debug", Name: "CrashOptions"}:                    {},
	{ImportPath: "runtime/metrics", Name: "Description"}:                   {},
	{ImportPath: "runtime/metrics", Name: "Sample"}:                        {},
	{ImportPath: "runtime/metrics", Name: "Value"}:                         {},
	{ImportPath: "runtime/trace", Name: "Region"}:                          {},
	{ImportPath: "runtime/trace", Name: "Task"}:                            {},
	{ImportPath: "strings", Name: "Reader"}:                                {},
	{ImportPath: "structs", Name: "HostLayout"}:                            {},
	{ImportPath: "syscall", Name: "Cmsghdr"}:                               {},
	{ImportPath: "syscall", Name: "Dirent"}:                                {},
	{ImportPath: "syscall", Name: "EpollEvent"}:                            {},
	{ImportPath: "syscall", Name: "FdSet"}:                                 {},
	{ImportPath: "syscall", Name: "Flock_t"}:                               {},
	{ImportPath: "syscall", Name: "Fsid"}:                                  {},
	{ImportPath: "syscall", Name: "ICMPv6Filter"}:                          {},
	{ImportPath: "syscall", Name: "IPMreq"}:                                {},
	{ImportPath: "syscall", Name: "IPMreqn"}:                               {},
	{ImportPath: "syscall", Name: "IPv6MTUInfo"}:                           {},
	{ImportPath: "syscall", Name: "IPv6Mreq"}:                              {},
	{ImportPath: "syscall", Name: "IfAddrmsg"}:                             {},
	{ImportPath: "syscall", Name: "IfInfomsg"}:                             {},
	{ImportPath: "syscall", Name: "Inet4Pktinfo"}:                          {},
	{ImportPath: "syscall", Name: "Inet6Pktinfo"}:                          {},
	{ImportPath: "syscall", Name: "InotifyEvent"}:                          {},
	{ImportPath: "syscall", Name: "Linger"}:                                {},
	{ImportPath: "syscall", Name: "NetlinkRouteRequest"}:                   {},
	{ImportPath: "syscall", Name: "NlAttr"}:                                {},
	{ImportPath: "syscall", Name: "NlMsgerr"}:                              {},
	{ImportPath: "syscall", Name: "NlMsghdr"}:                              {},
	{ImportPath: "syscall", Name: "PtraceRegs"}:                            {},
	{ImportPath: "syscall", Name: "RawSockaddr"}:                           {},
	{ImportPath: "syscall", Name: "RawSockaddrAny"}:                        {},
	{ImportPath: "syscall", Name: "RawSockaddrInet4"}:                      {},
	{ImportPath: "syscall", Name: "RawSockaddrInet6"}:                      {},
	{ImportPath: "syscall", Name: "RawSockaddrLinklayer"}:                  {},
	{ImportPath: "syscall", Name: "RawSockaddrNetlink"}:                    {},
	{ImportPath: "syscall", Name: "RawSockaddrUnix"}:                       {},
	{ImportPath: "syscall", Name: "Rlimit"}:                                {},
	{ImportPath: "syscall", Name: "RtAttr"}:                                {},
	{ImportPath: "syscall", Name: "RtGenmsg"}:                              {},
	{ImportPath: "syscall", Name: "RtMsg"}:                                 {},
	{ImportPath: "syscall", Name: "RtNexthop"}:                             {},
	{ImportPath: "syscall", Name: "Rusage"}:                                {},
	{ImportPath: "syscall", Name: "SockFilter"}:                            {},
	{ImportPath: "syscall", Name: "SockaddrInet4"}:                         {},
	{ImportPath: "syscall", Name: "SockaddrInet6"}:                         {},
	{ImportPath: "syscall", Name: "SockaddrLinklayer"}:                     {},
	{ImportPath: "syscall", Name: "SockaddrNetlink"}:                       {},
	{ImportPath: "syscall", Name: "SockaddrUnix"}:                          {},
	{ImportPath: "syscall", Name: "Stat_t"}:                                {},
	{ImportPath: "syscall", Name: "Statfs_t"}:                              {},
	{ImportPath: "syscall", Name: "SysProcIDMap"}:                          {},
	{ImportPath: "syscall", Name: "Sysinfo_t"}:                             {},
	{ImportPath: "syscall", Name: "TCPInfo"}:                               {},
	{ImportPath: "syscall", Name: "Termios"}:                               {},
	{ImportPath: "syscall", Name: "Timespec"}:                              {},
	{ImportPath: "syscall", Name: "Timeval"}:                               {},
	{ImportPath: "syscall", Name: "Timex"}:                                 {},
	{ImportPath: "syscall", Name: "Tms"}:                                   {},
	{ImportPath: "syscall", Name: "Ucred"}:                                 {},
	{ImportPath: "syscall", Name: "Ustat_t"}:                               {},
	{ImportPath: "syscall", Name: "Utimbuf"}:                               {},
	{ImportPath: "syscall", Name: "Utsname"}:                               {},
	{ImportPath: "testing", Name: "CoverBlock"}:                            {},
	{ImportPath: "text/scanner", Name: "Position"}:                         {},
	{ImportPath: "time", Name: "ParseError"}:                               {},
	{ImportPath: "unicode", Name: "CaseRange"}:                             {},
	{ImportPath: "unicode", Name: "Range16"}:                               {},
	{ImportPath: "unicode", Name: "Range32"}:                               {},
}
