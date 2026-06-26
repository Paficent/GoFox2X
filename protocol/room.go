package protocol

import "github.com/Paficent/GoFox2X/data"

type Room struct {
	ID                  int
	Name                string
	Type                string
	IsGame              bool
	IsHidden            bool
	IsPasswordProtected bool
	UserCount           int16
	MaxPlayers          int16
	Vars                []*data.GFSObject
}

func (r Room) ToSFSArray() *data.GFSArray {
	room := data.MakeGFSArray()
	room.AddInt(r.ID)
	room.AddUtfString(r.Name)
	room.AddUtfString(r.Type)
	room.AddBool(r.IsGame)
	room.AddBool(r.IsHidden)
	room.AddBool(r.IsPasswordProtected)
	room.AddShort(r.UserCount)
	room.AddShort(r.MaxPlayers)

	vars := data.MakeGFSArray()
	for _, v := range r.Vars {
		vars.AddSFSObject(v)
	}
	room.AddSFSArray(vars)

	return room
}

func RoomList(rooms ...Room) *data.GFSArray {
	list := data.MakeGFSArray()
	for _, r := range rooms {
		list.AddSFSArray(r.ToSFSArray())
	}
	return list
}
