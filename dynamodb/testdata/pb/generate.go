package pb

//go:generate protoc --gogo_out=.    events.proto
//go:generate protoc --event_out=.   events.proto

//go:generate protoc --gogo_out=.    commands.proto
//go:generate protoc --command_out=. commands.proto
