package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//TODO: Not so necessary of data driven testing. May update later.
func Test_addressTransactionLRU_putAddress(t *testing.T) {
	type fields struct {
		capability int
		dataMap    map[string]*addressTransactionNode
		head       *addressTransactionNode
		tail       *addressTransactionNode
	}
	type args struct {
		data []addressTransaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
		cases  []func() bool
	}{
		{
			name: "normal case 1",
			fields: fields{
				capability: 2,
			},
			args: args{
				data: []addressTransaction{
					{
						address:  "1",
						blockNum: 0,
					},
					{
						address:  "2",
						blockNum: 0,
					},
					{
						address:  "3",
						blockNum: 0,
					},
				},
			},
			want:  []string{},
			cases: []func() bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := newAddressTransactionLRU(tt.fields.capability)
			for _, d := range tt.args.data {
				this.putAddress(d)
			}
			assert.Equal(t, tt.fields.capability, this.size())
			assert.Equal(t, tt.fields.capability, len(this.allAddresses()))

			assert.Equal(t, (*addressTransaction)(nil), this.getAddress("1"))
			assert.NotEqual(t, (*addressTransaction)(nil), this.getAddress("2"))
			assert.Equal(t, "2", this.getAddress("2").address)
			assert.Equal(t, 0, this.getAddress("2").blockNum)

			assert.NotEqual(t, (*addressTransaction)(nil), this.getAddress("3"))

			sum := 0
			for node := this.head; node != nil; node = node.next {
				sum += 1
			}
			assert.Equal(t, 4, sum)
		})
	}
}

func Test_addressTransactionLRU_getAddress(t *testing.T) {
	type fields struct {
		capability int
		dataMap    map[string]*addressTransactionNode
		head       *addressTransactionNode
		tail       *addressTransactionNode
	}
	type args struct {
		addr string
		data []addressTransaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "normal case 1",
			fields: fields{
				capability: 1,
			},
			args: args{
				addr: "1",
				data: []addressTransaction{
					{
						address:  "1",
						blockNum: 0,
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := newAddressTransactionLRU(tt.fields.capability)
			for _, d := range tt.args.data {
				this.putAddress(d)
			}
			got := this.getAddress(tt.args.addr)
			assert.Equal(t, tt.args.data[0].address, got.address)
			assert.Equal(t, tt.args.data[0].blockNum, got.blockNum)
		})
	}
}

func Test_addressTransactionLRU_getAddressIn(t *testing.T) {
	type fields struct {
		capability int
		dataMap    map[string]*addressTransactionNode
		head       *addressTransactionNode
		tail       *addressTransactionNode
	}
	type args struct {
		addr string
		data []addressTransaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *addressTransaction
	}{
		{
			name: "normal case 1",
			fields: fields{
				capability: 2,
			},
			args: args{
				addr: "",
				data: []addressTransaction{
					{
						address:  "1",
						blockNum: 1,
					},
					{
						address:  "2",
						blockNum: 2,
					},
					{
						address:  "3",
						blockNum: 3,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := newAddressTransactionLRU(tt.fields.capability)
			for _, d := range tt.args.data {
				this.putAddress(d)
			}
			assert.Equal(t, (*addressTransaction)(nil), this.getAddressIn("1"))
			assert.NotEqual(t, (*addressTransaction)(nil), this.getAddressIn("2"))
			assert.NotEqual(t, (*addressTransaction)(nil), this.getAddressIn("3"))
		})
	}
}

func Test_addressTransactionLRU_size(t *testing.T) {
	type fields struct {
		capability int
		dataMap    map[string]*addressTransactionNode
		head       *addressTransactionNode
		tail       *addressTransactionNode
	}
	tests := []struct {
		name   string
		fields fields
		data   []addressTransaction
		want   int
	}{
		{
			name: "normal case 1",
			fields: fields{
				capability: 2,
			},
			data: []addressTransaction{
				{
					address:  "1",
					blockNum: 1,
				},
			},
			want: 1,
		},
		{
			name: "normal case 2",
			fields: fields{
				capability: 2,
			},
			data: []addressTransaction{
				{
					address:  "1",
					blockNum: 1,
				},
				{
					address:  "2",
					blockNum: 2,
				},
			},
			want: 2,
		},
		{
			name: "normal case 3",
			fields: fields{
				capability: 2,
			},
			data: []addressTransaction{
				{
					address:  "1",
					blockNum: 1,
				},
				{
					address:  "2",
					blockNum: 2,
				},
				{
					address:  "3",
					blockNum: 3,
				},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := newAddressTransactionLRU(tt.fields.capability)
			for _, d := range tt.data {
				this.putAddress(d)
			}
			assert.Equal(t, tt.want, this.size())
		})
	}
}

func Test_addressTransactionLRU_removeTail(t *testing.T) {
	type fields struct {
		capability int
		dataMap    map[string]*addressTransactionNode
		head       *addressTransactionNode
		tail       *addressTransactionNode
	}
	tests := []struct {
		name   string
		fields fields
		data   []addressTransaction
		want   int
	}{
		{
			name: "normal case 1",
			fields: fields{
				capability: 2,
			},
			data: []addressTransaction{
				{
					address:  "1",
					blockNum: 1,
				},
				{
					address:  "2",
					blockNum: 2,
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := newAddressTransactionLRU(tt.fields.capability)
			for _, d := range tt.data {
				this.putAddress(d)
			}
			this.removeTail()
			assert.Equal(t, tt.want, this.size())
			assert.Equal(t, (*addressTransaction)(nil), this.getAddress("1"))
			assert.NotEqual(t, (*addressTransaction)(nil), this.getAddress("2"))
		})
	}
}

func Test_addressTransactionLRU_allAddresses(t *testing.T) {
	type fields struct {
		capability int
		dataMap    map[string]*addressTransactionNode
		head       *addressTransactionNode
		tail       *addressTransactionNode
	}
	tests := []struct {
		name   string
		fields fields
		data   []addressTransaction
		want   []string
	}{
		{
			name: "normal case 1",
			fields: fields{
				capability: 2,
			},
			data: []addressTransaction{
				{
					address:  "1",
					blockNum: 1,
				},
				{
					address:  "2",
					blockNum: 2,
				},
			},
			want: []string{"2", "1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := newAddressTransactionLRU(tt.fields.capability)
			for _, d := range tt.data {
				this.putAddress(d)
			}
			got := this.allAddresses()
			assert.Equal(t, 2, len(got))
			assert.Equal(t, tt.want, got)
		})
	}
}
