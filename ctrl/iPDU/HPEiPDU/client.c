#ifdef HAVE_CONFIG_H
# include <config.h>
#endif

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <unistd.h>
#include <gnutls/gnutls.h>


/* A very basic TLS client, with X.509 authentication for HPE iPDU
  WARNING: THIS IS UNSAFE to use that tool when compile with TLS 1.0
  WARNING2: Certificate validation is disabled as most of our test PDU are embedding
  deprecated certificates

  command: iPDU on or off
  the port is set to CORE 1 - LOAD 6. If the test machine is on a different outlet
  that must be adapted withing the XML data sent to the iPDU
 */

#define MAX_BUF 4096
#define CAFILE "ca.pem"
#define ON "<?xml version=\"1.0\"?> \n\
	    <RIBCL VERSION=\"2.21\">\n\
		<LOGIN USER_LOGIN=\"admin\" PASSWORD=\"admin\">\n\
			  <SERVER_INFO MODE=\"write\">\n\
			  	  <SET_PDU_S_OUTLETCONTROL>\n\
				  	<ID CORE=\"1\" LOAD=\"6\" OUTLET=\"1\">\n\
						<SET_STATE VALUE=\"Y\"/>\n\
					</ID>\n\
				  </SET_PDU_S_OUTLETCONTROL>\n\
			  </SERVER_INFO>\n\
		</LOGIN>\n\
	    </RIBCL>"
#define OFF "<?xml version=\"1.0\"?> \n \
            <RIBCL VERSION=\"2.21\">\n\
                <LOGIN USER_LOGIN=\"admin\" PASSWORD=\"admin\">\n\
                          <SERVER_INFO MODE=\"write\">\n\
                                  <SET_PDU_S_OUTLETCONTROL>\n\
                                        <ID CORE=\"1\" LOAD=\"6\" OUTLET=\"1\">\n\
                                                <SET_STATE VALUE=\"N\"/>\n\
                                        </ID>\n\
                                  </SET_PDU_S_OUTLETCONTROL>\n\
                          </SERVER_INFO>\n\
                </LOGIN>\n\
            </RIBCL>"

int tcp_connect(void);
void tcp_close(int sd);

/* Connects to the peer and returns a socket
 * descriptor.
 */
extern int tcp_connect(void)
{
        const char *PORT = "50443";
        const char *SERVER = "10.4.0.199";
        long err, sd;
        struct sockaddr_in sa;

        /* connects to server
         */
        sd = socket(AF_INET, SOCK_STREAM, 0);

        memset(&sa, '\0', sizeof(sa));
        sa.sin_family = AF_INET;
        sa.sin_port = htons(atoi(PORT));
        inet_pton(AF_INET, SERVER, &sa.sin_addr);

        err = connect(sd, (struct sockaddr *) &sa, sizeof(sa));
        if (err < 0) {
                fprintf(stderr, "Connect error\n");
                exit(1);
        }

        return sd;
}

/* closes the given socket descriptor.
 */

extern void tcp_close(int sd)
{
        shutdown(sd, SHUT_RDWR);        /* no more receptions */
        close(sd);
}
extern int tcp_connect (void);
extern void tcp_close (int sd);

int
main (int argc, char *argv[])
{
  long ret, sd, ii;
  int command=-1;
  gnutls_session_t session;
  char buffer[MAX_BUF + 1];
  const char *err;

  if ( argc != 2 ) {
	  printf("please specify a parameter: iPDU <on/off>\n ");
	  exit(1);
  }
  if ( strcmp(argv[1], "on") == 0 ) {
	  command = 1;
  } else
	if ( strcmp(argv[1], "off") == 0 ) {
		command = 0;
	} else {
		printf("please specify a valid parameter: iPDU <on/off> \n");
		exit(1);
	}

  gnutls_certificate_credentials_t xcred;

  gnutls_global_init ();

  /* X509 stuff */
  gnutls_certificate_allocate_credentials (&xcred);

  /* sets the trusted cas file
   */

  /* Initialize TLS session 
   */
  gnutls_init (&session, GNUTLS_CLIENT);

  /* set priorities for iPDU */
  gnutls_set_default_priority (session);
  int parameters[1];
  parameters[0]=GNUTLS_CIPHER_3DES_CBC; //RSA_3DES_EDE_CBC_SHA1
  parameters[1]=0;
  gnutls_cipher_set_priority(session, (int *)(&parameters));
  parameters[0]=GNUTLS_TLS1_0;
  gnutls_protocol_set_priority(session, (int *)(&parameters));
  parameters[0]=GNUTLS_MAC_SHA1;
  gnutls_mac_set_priority(session, (int *)(&parameters));
  parameters[0]=GNUTLS_KX_RSA;
  gnutls_kx_set_priority(session, (int *)(&parameters));
  
  //  GNUTLS_SIGN_RSA_SHA
  /* put the x509 credentials to the current session
   */

  gnutls_credentials_set (session, GNUTLS_CRD_CERTIFICATE, xcred);

  /* connect to the peer
   */
  sd = tcp_connect ();

  gnutls_transport_set_ptr (session, (gnutls_transport_ptr_t) sd);

  /* Perform the TLS handshake
   */
  ret = gnutls_handshake (session);

  if (ret < 0)
    {
      fprintf (stderr, "*** Handshake failed\n");
      gnutls_perror (ret);
     // goto end;
    }
  else
    {
      printf ("- Handshake was completed\n");
    }
  if ( command == 1 )
	  gnutls_record_send (session, ON, strlen (ON));
  else
	  gnutls_record_send (session, OFF, strlen (OFF));

  ret = gnutls_record_recv (session, buffer, MAX_BUF);
  if (ret == 0)
    {
      printf ("- Peer has closed the TLS connection\n");
      goto end;
    }
  else if (ret < 0)
    {
      fprintf (stderr, "*** Error: %s\n", gnutls_strerror (ret));
      goto end;
    }

  printf ("- Received %ld bytes: ", ret);
  for (ii = 0; ii < ret; ii++)
    {
      fputc (buffer[ii], stdout);
    }
  fputs ("\n", stdout);

  gnutls_bye (session, GNUTLS_SHUT_RDWR);

end:

  tcp_close (sd);

  gnutls_deinit (session);

  gnutls_certificate_free_credentials (xcred);

  gnutls_global_deinit ();

  return 0;
}
