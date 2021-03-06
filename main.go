/*******************************************************************************
** @Author:					Major Tom - Sacré Studio <Major>
** @Email:					sacrestudioparis@gmail.com
** @Date:					Monday 03 September 2018 - 18:13:51
** @Filename:				main.go
**
** @Last modified by:		Tbouder
** @Last modified time:		Tuesday 14 April 2020 - 13:46:27
*******************************************************************************/

package			main

import			"log"
import			"os"
import			"net"
import			"crypto/tls"
import			"crypto/x509"
import			"io/ioutil"
import			"github.com/microgolang/logs"
import			"google.golang.org/grpc"
import			"google.golang.org/grpc/credentials"
import			"github.com/panghostlin/SDK/Members"
import			"github.com/panghostlin/SDK/Pictures"
import			"database/sql"
import			_ "github.com/lib/pq"

const	DEFAULT_CHUNK_SIZE = 64 * 1024
type	server		struct{}
type	sClients	struct {
	members		members.MembersServiceClient
	pictures	pictures.PicturesServiceClient
	albums		pictures.AlbumsServiceClient
}
var		PGR *sql.DB
var		bridges map[string](*grpc.ClientConn)
var		clients = &sClients{}

func	init() {
	os.Mkdir(`pictures`, os.ModePerm)
}

func	connectToDatabase() {
	username := os.Getenv("POSTGRE_USERNAME")
	password := os.Getenv("POSTGRE_PWD")
	host := os.Getenv("POSTGRE_URI")
	dbName := os.Getenv("POSTGRE_DB")
	connStr := "user=" + username + " password=" + password + " dbname=" + dbName + " host=" + host + " sslmode=disable"
	PGR, _ = sql.Open("postgres", connStr)

	PGR.Exec(`CREATE extension if not exists "uuid-ossp";`)
	/**************************************************************************
	**	Create a new class to allow the creation or array of uuid
	**************************************************************************/
	// PGR.Exec(`CREATE OPERATOR if not exists CLASS _uuid_ops DEFAULT FOR TYPE _uuid USING gin AS OPERATOR 1 &&(anyarray, anyarray), OPERATOR 2 @>(anyarray, anyarray), OPERATOR 3 <@(anyarray, anyarray), OPERATOR 4 =(anyarray, anyarray), FUNCTION 1 uuid_cmp(uuid, uuid), FUNCTION 2 ginarrayextract(anyarray, internal, internal), FUNCTION 3 ginqueryarrayextract(anyarray, internal, smallint, internal, internal, internal, internal), FUNCTION 4 ginarrayconsistent(internal, smallint, anyarray, integer, internal, internal, internal, internal), STORAGE uuid;`)

	PGR.Exec(`CREATE TABLE if not exists pictures(
		ID uuid NOT NULL DEFAULT uuid_generate_v4(),
		GroupID varchar NOT NULL,
		MemberID uuid NOT NULL,
		AlbumID uuid NULL,
		Size varchar NULL,
		Type varchar NULL,
		Name varchar NULL,
		EncryptionKey varchar NULL,
		EncryptionIV varchar NULL,
		Path varchar NULL,
		Width int NULL,
		Height int NULL,
		Weight float8 NULL,
		OriginalTime TIMESTAMP DEFAULT NOW(),

		CONSTRAINT pictures_pk PRIMARY KEY (ID)
	);`)
	PGR.Exec(`CREATE TABLE if not exists albums(
		ID uuid NOT NULL DEFAULT uuid_generate_v4(),
		MemberID uuid NOT NULL,
		Name varchar NULL,
		CoverPicture varchar NOT NULL DEFAULT '',
		NumberOfPictures int NOT NULL DEFAULT 0,
		CreationTime TIMESTAMP DEFAULT NOW(),
		CONSTRAINT albums_pk PRIMARY KEY (ID)
	);`)

	/**************************************************************************
	**	Create a function to update the album cover when a cover picture is
	**	removed
	**************************************************************************/
	PGR.Exec(`Create or replace
	function public.remove_cover() returns trigger language plpgsql as $function$ begin update albums set CoverPicture = '' where id = old.AlbumID and old.size = 'original' and CoverPicture = old.GroupID; return new; end; $function$ ;`)
	PGR.Exec(`create or replace function public.add_cover() returns trigger language plpgsql as $function$ begin update albums set CoverPicture = new.GroupID where id = new.AlbumID and new.Size = 'original' and CoverPicture = ''; return new; end; $function$ ;`)
	PGR.Exec(`CREATE trigger a_removeCover AFTER DELETE OR UPDATE on public.pictures for each row execute function remove_cover();`)
	PGR.Exec(`CREATE trigger a_insertCover AFTER INSERT OR UPDATE on public.pictures for each row execute function add_cover();`)

	/**************************************************************************
	**	Create a function to unset the albumID reference of a picture when the
	**	album is deleted
	**************************************************************************/
	PGR.Exec(`CREATE OR REPLACE FUNCTION public.unsetalbumid() RETURNS trigger LANGUAGE plpgsql AS $function$ begin update pictures set AlbumID = null where AlbumID = old.id; RETURN new; END; $function$ ;`)
	PGR.Exec(`CREATE trigger unsetalbumid AFTER DELETE on public.albums for each row execute function unsetalbumid();`)

	/**************************************************************************
	**	Create a function to update the album picture count on insert
	**	or update
	**************************************************************************/
	PGR.Exec(`CREATE OR REPLACE FUNCTION public.increase_album_pictures_count() RETURNS trigger LANGUAGE plpgsql AS $function$ begin UPDATE albums SET NumberOfPictures = NumberOfPictures - 1 WHERE id = old.AlbumID AND old.size = 'original';UPDATE albums SET NumberOfPictures = NumberOfPictures + 1 WHERE id = new.AlbumID AND new.size = 'original'; RETURN new; END; $function$;`)
	PGR.Exec(`CREATE trigger b_increasepictcount AFTER INSERT or UPDATE or DELETE on public.pictures for each row EXECUTE PROCEDURE public.increase_album_pictures_count()`)

	/**************************************************************************
	**	Create a function to update the member used storage on picture insert
	**	or delete
	**************************************************************************/
	PGR.Exec(`CREATE OR REPLACE FUNCTION public.f_update_member_used_storage() RETURNS trigger LANGUAGE plpgsql AS $function$ BEGIN IF(TG_OP = 'DELETE') THEN update members set usedStorage = usedStorage - old.weight where ID = old.memberID and old.size = 'original'; update members set fullusedStorage = fullusedStorage - old.weight where ID = old.memberID; return old; ELSIF (TG_OP = 'INSERT') THEN update members set usedStorage = usedStorage + new.weight where ID = new.memberID and new.size = 'original'; update members set fullusedStorage = fullusedStorage + new.weight where ID = new.memberID; return new; END IF; END; $function$ ;`)
	PGR.Exec(`create trigger c_update_member_used_storage before insert or delete on public.pictures for each row execute function f_update_member_used_storage();`)
	logs.Success(`Connected to DB - Localhost`)
}
func	bridgeInsecureMicroservice(serverName string, clientMS string) (*grpc.ClientConn) {
	logs.Warning("Using insecure connection")
	conn, err := grpc.Dial(serverName, grpc.WithInsecure())
    if err != nil {
		logs.Error("Did not connect", err)
		return nil
	}

	if (clientMS == `members`) {
		clients.members = members.NewMembersServiceClient(conn)
	} else if (clientMS == `pictures`) {
		clients.pictures = pictures.NewPicturesServiceClient(conn)
		clients.albums = pictures.NewAlbumsServiceClient(conn)
	}

	return conn
}
func	bridgeMicroservice(serverName string, clientMS string) (*grpc.ClientConn){
	crt := `/env/client.crt`
    key := `/env/client.key`
	caCert  := `/env/ca.crt`

    // Load the client certificates from disk
    certificate, err := tls.LoadX509KeyPair(crt, key)
    if err != nil {
		logs.Warning("Did not connect: " + err.Error())
		return bridgeInsecureMicroservice(serverName, clientMS)
    }

    // Create a certificate pool from the certificate authority
    certPool := x509.NewCertPool()
    ca, err := ioutil.ReadFile(caCert)
    if err != nil {
		logs.Warning("Did not connect: " + err.Error())
		return bridgeInsecureMicroservice(serverName, clientMS)
    }

    // Append the certificates from the CA
    if ok := certPool.AppendCertsFromPEM(ca); !ok {
		logs.Warning("Did not connect: " + err.Error())
		return bridgeInsecureMicroservice(serverName, clientMS)
    }

    creds := credentials.NewTLS(&tls.Config{
        ServerName:   serverName, // NOTE: this is required!
        Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
		InsecureSkipVerify: true,
    })

    // Create a connection with the TLS credentials
	conn, err := grpc.Dial(serverName, grpc.WithTransportCredentials(creds))
    if err != nil {
		logs.Warning("Did not connect: " + err.Error())
		return bridgeInsecureMicroservice(serverName, clientMS)
	}

	if (clientMS == `members`) {
		clients.members = members.NewMembersServiceClient(conn)
	} else if (clientMS == `pictures`) {
		clients.pictures = pictures.NewPicturesServiceClient(conn)
	}

	return conn
}
func	serveInsecureMicroservice() {
    lis, err := net.Listen(`tcp`, `:8012`)
    if err != nil {
		log.Fatalf("Failed to listen: %v", err)
    }

	srv := grpc.NewServer()
	pictures.RegisterPicturesServiceServer(srv, &server{})
	pictures.RegisterAlbumsServiceServer(srv, &server{})
	logs.Success(`Running on port: :8012`)
	if err := srv.Serve(lis); err != nil {
		logs.Error(err)
		log.Fatalf("failed to serve: %v", err)
	}
}
func	serveMicroservice() {
	crt := `/env/server.crt`
    key := `/env/server.key`
	caCert  := `/env/ca.crt`
	
	certificate, err := tls.LoadX509KeyPair(crt, key)
    if err != nil {
		logs.Warning("could not load server key pair : " + err.Error())
		logs.Warning("Using insecure connection")
		serveInsecureMicroservice()
    }

    // Create a certificate pool from the certificate authority
    certPool := x509.NewCertPool()
    ca, err := ioutil.ReadFile(caCert)
    if err != nil {
        log.Fatalf("could not read ca certificate: %s", err)
    }

    // Append the client certificates from the CA
    if ok := certPool.AppendCertsFromPEM(ca); !ok {
        log.Fatalf("failed to append client certs")
    }

    // Create the channel to listen on
    lis, err := net.Listen(`tcp`, `:8012`)
    if err != nil {
		log.Fatalf("Failed to listen: %v", err)
    }

    // Create the TLS credentials
    creds := credentials.NewTLS(&tls.Config{
    	ClientAuth:   tls.RequireAndVerifyClientCert,
    	Certificates: []tls.Certificate{certificate},
    	ClientCAs:    certPool,
	})

    // Create the gRPC server with the credentials
    srv := grpc.NewServer(grpc.Creds(creds))

	// Register the handler object
	pictures.RegisterPicturesServiceServer(srv, &server{})
	pictures.RegisterAlbumsServiceServer(srv, &server{})

    // Serve and Listen
	logs.Success(`Running on port :8012`)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func	main()	{
	connectToDatabase()
	serveMicroservice()
}
