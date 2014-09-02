#include <string.h>
#include <glib.h>
#include <glib/gprintf.h>
#include <openssl/bio.h>
#include <openssl/x509.h>
#include <openssl/evp.h>
#include <openssl/sha.h>
#include <openssl/ec.h>

#define SALT "dscuss-proof-of-work"
#define REQUIRED_ZERO_NUM 25
#define g_htonll(val) (GUINT64_TO_BE (val))


struct DscussHash
{
  unsigned char digest[SHA512_DIGEST_LENGTH];
};


int
hash_get_bit (const struct DscussHash* hash,
              unsigned int bit)
{
  g_assert (bit < 8 * sizeof (struct DscussHash));
  return (((unsigned char *) hash)[bit >> 3] & (1 << (bit & 7))) > 0;
}


/**
 * Count the leading zeroes in hash.
 *
 * @param hash to count leading zeros in
 * @return the number of leading zero bits.
 */
static unsigned int
count_leading_zeroes (const struct DscussHash* hash)
{
  unsigned int hash_count;

  hash_count = 0;
  while ((0 == hash_get_bit (hash, hash_count)))
    hash_count++;
  return hash_count;
}


int
main (int argc, char* argv[])
{
  EC_KEY* eckey = NULL;
  BIO* bio = NULL;
  gsize keylen;
  gchar *to_hash = NULL;
  unsigned char digest[SHA512_DIGEST_LENGTH];
  char digest_hex_str[2 * SHA512_DIGEST_LENGTH + 1];
  unsigned int i = 0;

  
  eckey = EC_KEY_new_by_curve_name (NID_secp224r1);
  if (NULL == eckey)
    {
      g_error ("Failed to create new EC key");
      goto out;
    }

  if (EC_KEY_generate_key (eckey) != 1)
    {
      g_error ("Failed to generate EC key");
      goto out;
    }

  EC_KEY_set_asn1_flag(eckey, OPENSSL_EC_NAMED_CURVE);

  bio = BIO_new (BIO_s_mem ());
  if (bio == NULL)
    {
      g_warning ("Failed to create new BIO");
      goto out;
    }

  if (i2d_EC_PUBKEY_bio (bio, eckey) != 1)
    {
      g_warning ("Failed to write public key to BIO");
      goto out;
    }

  keylen = BIO_pending (bio);
  to_hash = g_malloc0 (keylen + sizeof (guint64));
  if (BIO_read (bio, to_hash, keylen) <= 0)
    {
      g_warning ("Failed to read key digest from BIO");
      goto out;
    }

  g_printf ("EC public key length = %lu\n", keylen);

  guint64 proof = 0;
  guint64 counter = proof;
  g_assert(argc > 1);
  gint required_zero_num = atoi(argv[1]);
  g_printf ("Required zero num = %d\n", required_zero_num);
  guint iteration = 0;
  for (counter = proof;
       counter < G_MAXUINT64;
       counter++)
    {
      guint64 ncounter = g_htonll (counter);
      memcpy (&to_hash[keylen], &ncounter, sizeof (guint64));
      PKCS5_PBKDF2_HMAC (to_hash,
                         keylen + SHA512_DIGEST_LENGTH,
                         (const unsigned char *) SALT,
                         strlen (SALT),
                         1, // iterations 
                         EVP_sha512 (),
                         SHA512_DIGEST_LENGTH,
                         digest);
      if (count_leading_zeroes ((const struct DscussHash*) digest) >= required_zero_num)
        {
          proof = counter;
          g_printf ("Proof of work found: = %lu\n", proof);
          memset (digest_hex_str, 0, sizeof (digest_hex_str));
          for (i = 0; i < sizeof (digest); i++)
            sprintf(digest_hex_str + (i * 2), "%02x", 255 & digest[i]);
          g_printf ("%s\n", digest_hex_str);
          g_printf ("Number of leading zeros = %u\n",
                    count_leading_zeroes ((const struct DscussHash*) digest));
          if (iteration == 9)
            break;
          else
            iteration++;
        }
    }



out:
  if (bio != NULL)
    BIO_free_all (bio);

  if (eckey != NULL)
    EC_KEY_free (eckey);

  if (to_hash != NULL)
    g_free (to_hash);

  return 0;
}
