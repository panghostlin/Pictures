/*******************************************************************************
** @Author:					Thomas Bouder <Tbouder>
** @Email:					Tbouder@protonmail.com
** @Date:					Tuesday 07 January 2020 - 16:37:26
** @Filename:				keyBridge.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Saturday 15 February 2020 - 14:12:46
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
func	encryptPictureSender(stream keys.KeysService_EncryptPictureClient, toEncrypt []byte, memberID string) (error) {
	fileSize := len(toEncrypt)
	chnk := &keys.EncryptPictureRequest{MemberID: memberID}

	for currentByte := 0; currentByte < fileSize; currentByte += DEFAULT_CHUNK_SIZE {
		if currentByte + DEFAULT_CHUNK_SIZE > fileSize {
			chnk.Chunk = toEncrypt[currentByte:fileSize]
		} else {
			chnk.Chunk = toEncrypt[currentByte : currentByte + DEFAULT_CHUNK_SIZE]
		}
		if err := stream.Send(chnk); err != nil {
			return err
		}
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}
	return nil
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
	}
}


/******************************************************************************
**	decryptPictureSender
**	Helper to DecryptPicture
**	Open a sending stream, to send the data to decrypt, with the memberID
**	asking for decryption and the picture decryption key.
******************************************************************************/
func	decryptPictureSender(stream keys.KeysService_DecryptPictureClient, toDecrypt []byte, memberID, key, hashKey string) (error) {
	fileSize := len(toDecrypt)
	req := &keys.DecryptPictureRequest{MemberID: memberID, Key: key, HashKey: hashKey}

	for currentByte := 0; currentByte < fileSize; currentByte += DEFAULT_CHUNK_SIZE {
		if currentByte + DEFAULT_CHUNK_SIZE > fileSize {
			req.Chunk = toDecrypt[currentByte:fileSize]
		} else {
			req.Chunk = toDecrypt[currentByte : currentByte + DEFAULT_CHUNK_SIZE]
		}
		if err := stream.Send(req); err != nil {
			logs.Error(err)
			stream.CloseSend()
			return err
		}
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}
	return nil
}
/******************************************************************************
**	decryptPictureReceiver
**	Helper to DecryptPicture
**	Open a receiving stream, waiting to get the get the full decrypted picture
******************************************************************************/
func	decryptPictureReceiver(stream keys.KeysService_DecryptPictureClient) (*keys.DecryptPictureResponse, error) {
	resp := &keys.DecryptPictureResponse{}

	for {
		select {
			case <-stream.Context().Done():
				return nil, stream.Context().Err()
			default:
		}

		receiver, err := stream.Recv()
		if err == io.EOF {
			stream.Context().Done()
			return resp, nil
		}
		if err != nil {
			// logs.Error("receive error : " + err.Error())
			return nil, err
		}

		resp.Chunk = append(resp.GetChunk(), receiver.GetChunk()...)
	}
}
/******************************************************************************
**	DecryptPicture
**	Bridge function, calling the Key microservice to decrypt a specific data
**	according to the memberID private key and the image encryption key.
******************************************************************************/
func	DecryptPicture(memberID string, toDecrypt []byte, key, hashKey string) ([]byte, error) {
	var stream keys.KeysService_DecryptPictureClient
	
	stream, err := clients.keys.DecryptPicture(context.Background())
	if (err != nil) {
		logs.Error("Fail to init stream", err)
		return nil, err
	}
	defer stream.Context().Done()

	/**************************************************************************
	**	1. Send the data to encrypt to the Key microservice
	**************************************************************************/
	err = decryptPictureSender(stream, toDecrypt, memberID, key, hashKey)
	if (err != nil) {
		return nil, err
	}

	/**************************************************************************
	**	2. Receive message from the stream, AKA from the EncryptMessage srv
	**	on the microservice.
	**	Here, we will get our answers
	**************************************************************************/
	response, err := decryptPictureReceiver(stream)
	if (err != nil) {
		return nil, err
	}

	/**************************************************************************
	**	3. Returns the elements
	**************************************************************************/
	return response.GetChunk(), nil
}
