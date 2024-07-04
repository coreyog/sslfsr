set windows-shell := ["powershell", "-c"]

default:
  @just --list

solver4:
  @cd cmd/solver4; go run .

solver4-debug:
  @cd cmd/solver4; go generate ./...; ./solver4 --wfd

solver8:
  @cd cmd/solver8; go run .

solver8-debug:
  @cd cmd/solver8; go generate ./...; ./solver8 --wfd

solver16:
  @cd cmd/solver16; go run .

solver16-debug:
  @cd cmd/solver16; go generate ./...; ./solver16 --wfd

test:
  @go test ./... -count=1

run-all-solvers: solver4 solver8 solver16

everything: test run-all-solvers
