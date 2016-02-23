package module

type Note struct {
	key int
	instrument int
	period int
	effect int
	parameter int
}

//Info for each note:
//
//_____byte 1_____   byte2_    _____byte 3_____   byte4_
///                \ /      \  /                \ /      \
//0000          0000-00000000  0000          0000-00000000
//
//Upper four    12 bits for    Lower four    Effect command.
//bits of sam-  note period.   bits of sam-
//ple number.                  ple number.

func (n *Note) Load(data []byte) error {

	n.period = int( ( (data[0] & 0xF) << 8) | ( data[1] & 0xFF) )
	n.instrument = int(( data[0] & 0x10 ) | ( ( data[2] & 0xF0) >> 4))
	n.effect = int(data[2] & 0xFF)
	n.parameter = int(data[3] & 0xFF)

	return nil
}