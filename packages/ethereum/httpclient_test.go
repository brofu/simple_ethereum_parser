// The test cases in this file are more like integration test.
// Need to start the testserver to run the cases in this file
// go run cmd/testserver/main.go

package ethereum

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEthJsonRpcClient_EthGetCurrentBlockNumber(t *testing.T) {
	type args struct {
		context context.Context
		req     *EthGetCurrentBlockNumberRequest
	}
	tests := []struct {
		name    string
		this    *EthJsonRpcClient
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "normal case 1",
			args: args{
				context: context.Background(),
				req: &EthGetCurrentBlockNumberRequest{
					RequestId: "1024",
				},
			},
			want:    int(time.Now().Unix()) - 1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &EthJsonRpcClient{
				entryPoint: testEntryPoint,
			}
			got, err := this.EthGetCurrentBlockNumber(tt.args.context, tt.args.req)
			assert.Equal(t, nil, err)
			fmt.Println("flag", tt.want, got)
			assert.Equal(t, true, tt.want <= got)
		})
	}
}

func TestEthJsonRpcClient_EthGetCurrentTransactionsByAddress(t *testing.T) {
	type fields struct {
		entryPoint string
	}
	type args struct {
		context context.Context
		req     *EthGetCurrentTransactionsByAddressRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "normal case 1",
			args: args{
				context: context.Background(),
				req: &EthGetCurrentTransactionsByAddressRequest{
					FromBlock:   "0x64",
					ToBlock:     "0xc8",
					FromAddress: "0xffff",
					ToAddress:   "0xffff",
					RequestId:   "1024",
				},
			},
			want:    2,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &EthJsonRpcClient{
				entryPoint: testEntryPoint,
			}

			got, err := this.EthGetCurrentTransactionsByAddress(tt.args.context, tt.args.req)
			assert.Equal(t, nil, err)
			assert.NotEqual(t, nil, got)
			assert.Equal(t, tt.want, len(got))
		})
	}
}
