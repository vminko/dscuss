/**
 * This file is part of Dscuss.
 * Copyright (C) 2014  Vitaly Minko
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * Additional permission under GNU GPL version 3 section 7
 *
 * If you modify this program, or any covered work, by linking or
 * combining it with the OpenSSL project's OpenSSL library (or a
 * modified version of that library), containing parts covered by the
 * terms of the OpenSSL or SSLeay licenses, the copyright holders
 * grants you additional permission to convey the resulting work.
 * Corresponding Source for a non-source form of such a combination
 * shall include the source code for the parts of OpenSSL used as well
 * as that of the covered work.
 */

/**
 * @file crypto.h  Elliptic curve cryptography with OpenSSL.
 */


#ifndef DSCUSS_CRYPTO_H
#define DSCUSS_CRYPTO_H

#include <glib.h>
#include <openssl/ec.h>
#include <openssl/pem.h>


#ifdef __cplusplus
extern "C" {
#endif


// Depends on CURVE, use ECDSA_sign to calculate it.
#define DSCUSS_CRYPTO_SIGNATURE_SIZE 64


typedef EC_KEY DscussPrivateKey;
typedef EC_POINT DscussPublicKey;

/* DER-encoded ECDSA signature. */
struct DscussSignature
{
  gchar s[DSCUSS_CRYPTO_SIGNATURE_SIZE];
};


/**
 * Generates new private key.
 *
 * @return new private key in case of success, or @c NULL on error.
 */
DscussPrivateKey*
dscuss_crypto_private_key_new ();

/**
 * Frees memory allocated for a private key.
 *
 * @param privkey   Private key to free.
 */
void
dscuss_crypto_private_key_free (DscussPrivateKey* privkey);

/**
 * Writes a private key to a file.
 *
 * @param privkey   Private key to store.
 * @param filename  Name of the file to write in.
 *
 * @return @c TRUE in case of success, or @c FALSE on error.
 */
gboolean
dscuss_crypto_private_key_write (const DscussPrivateKey* privkey,
                                 const gchar* filename);

/**
 * Reads private key from a file.
 *
 * @param filename  Name of the file to read from.
 *
 * @return  Private key in case, or @c NULL on error.
 */
DscussPrivateKey*
dscuss_crypto_private_key_read (const gchar* filename);

/**
 * Initialize private key by reading it from a file. If the files does not
 * exist, create a new private key and write it to the file.
 *
 * @param filename  Name of the file to read from or to store to if it does
 *                  not exist.
 *
 * @return  Private key in case, or @c NULL on error.
 */
DscussPrivateKey*
dscuss_crypto_private_key_init (const gchar* filename);

/**
 * Extracts the public key from the given private key.
 *
 * @param privkey  The private key.
 *
 * @return The public key in case of success, or @c NULL on error.
 */
const DscussPublicKey*
dscuss_crypto_private_key_get_public (const DscussPrivateKey* privkey);

/**
 * Encodes a public key into DER format.
 *
 * @param pubkey     Public key to encode.
 * @param digest     Where to store serialized key.
 * @param digest_len Where to store length of the @a digest.
 *
 * @return @c TRUE in case of success, or @c FALSE on error.
 */
gboolean
dscuss_crypto_public_key_to_der (const DscussPublicKey* pubkey,
                                 gchar** digest,
                                 gsize* digest_len);

/**
 * Decodes a public key from DER format.
 *
 * @param digest     Address of the serialized key.
 * @param digest_len Length of the @a digest.
 *
 * @return Decoded public key or @c NULL on error.
 */
DscussPublicKey*
dscuss_crypto_public_key_from_der (const gchar* digest,
                                   gsize digest_len);

/**
 * Creates a copy of a public key.
 *
 * @param pubkey  Public key to copy.
 *
 * @return  Newly created public key identical to @a pubkey.
 */
DscussPublicKey*
dscuss_crypto_public_key_copy (const DscussPublicKey*);

/**
 * Frees memory allocated for a public key.
 *
 * @param pubkey   Public key to free.
 */
void
dscuss_crypto_public_key_free (DscussPublicKey* pubkey);

/**
 * Sign a digest.
 *
 * @param digest         Address of the digest to sign.
 * @param digest_len     Length of the @a digest.
 * @param privkey        Private key to use for signing.
 * @param signature      Where to write the signature (output parameter).
 * @param signature_len  Length of the signature written (output parameter).
 *
 * @return @c TRUE in case of success, or @c FALSE on error.
 */
gboolean
dscuss_crypto_sign (const gchar* digest,
                    gsize digest_len,
                    const DscussPrivateKey* privkey,
                    struct DscussSignature* signature,
                    gsize* signature_len);

/**
 * Verify a signature.
 *
 * @param digest         Address of the digest to verify.
 * @param digest_len     Length of the @a digest.
 * @param pubkey         Public key of the signer.
 * @param signature      Signature to verify.
 * @param signature_len  Length of the signature.
 *
 * @return @c TRUE if signature is valid, or @c FALSE on error.
 */
gboolean
dscuss_crypto_verify (const gchar* digest,
                      gsize digest_len,
                      const DscussPublicKey* pubkey,
                      const struct DscussSignature* signature,
                      gsize signature_len);


#ifdef __cplusplus
}
#endif

#endif /* DSCUSS_CRYPTO_H */
