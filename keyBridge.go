/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Tuesday 07 January 2020 - 16:37:26
** @Filename:				keyBridge.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Monday 10 February 2020 - 11:56:39
*******************************************************************************/

package			main

import			"io"
import			"context"
import			"github.com/microgolang/logs"
import			"github.com/panghostlin/SDK/Keys"

/******************************************************************************
**	encryptPictureSender
**	Helper to EncryptPicture
**	Open a sending stream, to send the data to encrypt, with the memberID
**	asking for decryption.
******************************************************************************/
func	encryptPictureSender(stream keys.KeysService_EncryptPictureClient, toEncrypt []byte, memberID string) (bool, error) {
	fileSize := len(toEncrypt)
	chnk := &keys.EncryptPictureRequest{MemberID: memberID}

	for currentByte := 0; currentByte < fileSize; currentByte += DEFAULT_CHUNK_SIZE {
		if currentByte + DEFAULT_CHUNK_SIZE > fileSize {
			chnk.Chunk = toEncrypt[currentByte:fileSize]
		} else {
			chnk.Chunk = toEncrypt[currentByte : currentByte + DEFAULT_CHUNK_SIZE]
		}
		if err := stream.Send(chnk); err != nil {
			return false, err
		}
	}

	if err := stream.CloseSend(); err != nil {
		return false, err
	}
	return true, nil
}
/******************************************************************************
**	encryptPictureReceiver
**	Helper to EncryptPicture
**	Open a receiving stream, waiting to get the get the full encrypted picture
******************************************************************************/
func	encryptPictureReceiver(stream keys.KeysService_EncryptPictureClient) (*keys.EncryptPictureResponse, error) {
	resp := &keys.EncryptPictureResponse{}

	for {
		select {
			case <-stream.Context().Done():
				return nil, stream.Context().Err()
			default:
		}

		receiver, err := stream.Recv()
		if err == io.EOF {
			return resp, nil
		}
		if err != nil {
			logs.Error("receive error : ", err)
			continue
		}

		resp.Chunk = append(resp.GetChunk(), receiver.GetChunk()...)
		resp.Key = receiver.GetKey()
		resp.Success = receiver.GetSuccess()
	}
}


/******************************************************************************
**	decryptPictureSender
**	Helper to DecryptPicture
**	Open a sending stream, to send the data to decrypt, with the memberID
**	asking for decryption and the picture decryption key.
******************************************************************************/
func	decryptPictureSender(stream keys.KeysService_DecryptPictureClient, toDecrypt []byte, memberID, key, hashKey string) (bool, error) {
	fileSize := len(toDecrypt)
	req := &keys.DecryptPictureRequest{MemberID: memberID, Key: key, HashKey: hashKey}

	for currentByte := 0; currentByte < fileSize; currentByte += DEFAULT_CHUNK_SIZE {
		if currentByte + DEFAULT_CHUNK_SIZE > fileSize {
			req.Chunk = toDecrypt[currentByte:fileSize]
		} else {
			req.Chunk = toDecrypt[currentByte : currentByte + DEFAULT_CHUNK_SIZE]
		}
		if err := stream.Send(req); err != nil {
			logs.Info(`CRASH`)
			// logs.Error(err)
			stream.CloseSend()
			return false, err
		}
	}

	if err := stream.CloseSend(); err != nil {
		return false, err
	}
	return true, nil
}
/******************************************************************************
**	decryptPictureReceiver
**	Helper to DecryptPicture
**	Open a receiving stream, waiting to get the get the full decrypted picture
******************************************************************************/
func	decryptPictureReceiver(stream keys.KeysService_DecryptPictureClient) (*keys.DecryptPictureResponse, error) {
	blob := make([]byte, 0)
	success := false
	resp := &keys.DecryptPictureResponse{}

	for {
		select {
			case <-stream.Context().Done():
				return nil, stream.Context().Err()
			default:
		}

		receiver, err := stream.Recv()
		if err == io.EOF {
			resp.Chunk = blob
			resp.Success = success
			stream.Context().Done()
			return resp, nil
		}
		if err != nil {
			// logs.Error("receive error : " + err.Error())
			return nil, err
		}

		blob = append(blob, receiver.GetChunk()...)
		success = receiver.GetSuccess()
	}
}
/******************************************************************************
**	DecryptPicture
**	Bridge function, calling the Key microservice to decrypt a specific data
**	according to the memberID private key and the image encryption key.
******************************************************************************/
func	DecryptPicture(memberID string, toDecrypt []byte, key, hashKey string) (bool, []byte, error) {
	var stream keys.KeysService_DecryptPictureClient
	
	stream, err := clients.keys.DecryptPicture(context.Background())
	if (err != nil) {
		logs.Error("Fail to init stream", err)
		return false, nil, err
	}
	defer stream.Context().Done()

	/**************************************************************************
	**	1. Send the data to encrypt to the Key microservice
	**************************************************************************/
	isSuccess, err := decryptPictureSender(stream, toDecrypt, memberID, key, hashKey)
	if (err != nil || !isSuccess) {
		return false, nil, err
	}

	/**************************************************************************
	**	2. Receive message from the stream, AKA from the EncryptMessage srv
	**	on the microservice.
	**	Here, we will get our answers
	**************************************************************************/
	response, err := decryptPictureReceiver(stream)
	if (err != nil || !isSuccess) {
		return false, nil, err
	}

	/**************************************************************************
	**	3. Returns the elements
	**************************************************************************/
	return response.GetSuccess(), response.GetChunk(), nil
}
