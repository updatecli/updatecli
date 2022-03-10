package sign

import (
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp/errors"
)

type DataSet []Data

type Data struct {
	SigningKey    string
	Passphrase    string
	Enabled       bool
	ExpectedError error
}

var (
	dataset = DataSet{
		{
			SigningKey: `
-----BEGIN PGP PRIVATE KEY BLOCK-----

lQOsBGIqZEEBCACOnlRGd40w2SvqxFe0FRtBhruQIHLWpWAkYxz6otNRuCAkK7w1
GwPvFlxpYuR1Yb2X4Ex1skXGbCnb7yTwgJIKXV6btagA8MDYnonzGudnNQE8vKr3
IKRp5fDHYjIrml4jmyLiDiGvGhjQLwlB6Xc+qLZZVGiHy0YKJNWZiCF5HfJhszB5
cWeUgXwXfqLXYbhfn4qSRUKN0jcpnUfI+UWCAaH3vuodo7q/8rm4ESh6ioiCDe3M
a0+jMPa+EaLUv6Rlye+UBzj5xLWXEMgJwJVFJKXtH1mjGCM6H7herVEyChZmjXo9
RFnTz6s4Ohh9YuCbiRDQlueghp58oVJeo42fABEBAAH/AwMCFmrvPQSYi65g9w4O
cPSntz8UAm5sVksLxP7ScjEAw1KVel+bEcj+TOfY/Io+H05+X080PBdffbmsFRIK
BRoXb0BfCgpY7cDHgPuYKXcxfzJfom3MxmjFulrXI0Ia3ALWbSarwSAWFDqyoSE2
0i8ZTin/RNAuFEAwJ+9jrj/IjY8R2aIe4i4TZ8llmAKBQdTITkZQ1c1pZbZNYiWG
LVMAiM4RmYg21DVwoAcRFWOnjYgsYULyIz9GW9KsQMSZ/9oU2axT6BBQ0EcMIda3
fCoWg17EGC/g55rz36MXANIeaLf5DB1GEi+HMy6geq81fFChQt/DwopIWUImUZBL
6JYLqkTknbu1ZByYaZeHq+sWpa1dlCa38ToFjgYonI9X8QS5rG/c6nrplCC+9gWB
0HTlHSFK72unFhuyrBMg0eoKJa1btd5lvUdulM51+MfAvk3KcIn7iFIdxsZC/jW4
4GLdjihwzqKVH9DVRGdip1iNEjdDzpZ5O+2I+QuHGMRLnPN2dqTOW+dxtSfOBScf
vyozYlQBjT/sVQM5k6tWbiAagNwQLPSZbzSwNTrf0k4yakU051hf80TS3qTdmkCF
MDWxMlebESwcDoQOVyq7pLNyYN3Y5WwiBlaZZLAD71uztfAPMJQ+DWLq/nhLuE5v
xMTeTBtlN5qy2/LPjrqMW5JoMWzOJO3IZdP8Vo5d4LvC29Cy7UOgEwYsBJ0KDo/Z
Uz8uWucsPevJ5fEmpa6EkWlSm7adky2pMCPdxfdMT+RjYBKShr8NSsXe4ukbXakt
mHYJcqCbKlUy3J1/U1g07JUtv6jKEjUPUdKPqy0iXTglmdeOs5OWfffX9TITNxdQ
31RZjCVwIGiuszfdQ572qGPZw49SitoTMxJshMJEY7QLZm9vQGJhci5jb22JARwE
EAECAAYFAmIqZEEACgkQXAIYP4CH4zaoNQf/cSsNjUq84Frd1l96o3stxw/k7DGa
x1SvQ6T4zTlTndWZirnPv3OncrtzjpXjCFbymxBKhFKgrTA5hgT/Gm0z/BQWEGqf
GtiIcTnoPIJfrxGXYqqaMVMzFUIha3vfqf4DqmOcZIebUmOp926uWKnp7AMFt2Qs
1uhGIq5gxf44zKp3dKfLTU0GaqsGx8F0Hfc8uguXhKB8OdkhmwEwZMdFvTg4y8px
akA1wLgb4M2/j876jL16SacsmrOfD/f9OygPCUJWOPf83frmcJHrJJGeJ2G+V4Nv
yxxawXzHgc4q6+YvBjgCn3LWojXqh/07nRLiRNfFNX5DKKSa7AUE4ET4Mg==
=YcmB
-----END PGP PRIVATE KEY BLOCK-----
`,
			Passphrase:    "abcd123",
			ExpectedError: nil,
		},
		{
			SigningKey: `
-----BEGIN PGP PRIVATE KEY BLOCK-----

lQOsBGIqZEEBCACOnlRGd40w2SvqxFe0FRtBhruQIHLWpWAkYxz6otNRuCAkK7w1
GwPvFlxpYuR1Yb2X4Ex1skXGbCnb7yTwgJIKXV6btagA8MDYnonzGudnNQE8vKr3
IKRp5fDHYjIrml4jmyLiDiGvGhjQLwlB6Xc+qLZZVGiHy0YKJNWZiCF5HfJhszB5
cWeUgXwXfqLXYbhfn4qSRUKN0jcpnUfI+UWCAaH3vuodo7q/8rm4ESh6ioiCDe3M
a0+jMPa+EaLUv6Rlye+UBzj5xLWXEMgJwJVFJKXtH1mjGCM6H7herVEyChZmjXo9
RFnTz6s4Ohh9YuCbiRDQlueghp58oVJeo42fABEBAAH/AwMCFmrvPQSYi65g9w4O
cPSntz8UAm5sVksLxP7ScjEAw1KVel+bEcj+TOfY/Io+H05+X080PBdffbmsFRIK
BRoXb0BfCgpY7cDHgPuYKXcxfzJfom3MxmjFulrXI0Ia3ALWbSarwSAWFDqyoSE2
0i8ZTin/RNAuFEAwJ+9jrj/IjY8R2aIe4i4TZ8llmAKBQdTITkZQ1c1pZbZNYiWG
LVMAiM4RmYg21DVwoAcRFWOnjYgsYULyIz9GW9KsQMSZ/9oU2axT6BBQ0EcMIda3
fCoWg17EGC/g55rz36MXANIeaLf5DB1GEi+HMy6geq81fFChQt/DwopIWUImUZBL
6JYLqkTknbu1ZByYaZeHq+sWpa1dlCa38ToFjgYonI9X8QS5rG/c6nrplCC+9gWB
0HTlHSFK72unFhuyrBMg0eoKJa1btd5lvUdulM51+MfAvk3KcIn7iFIdxsZC/jW4
4GLdjihwzqKVH9DVRGdip1iNEjdDzpZ5O+2I+QuHGMRLnPN2dqTOW+dxtSfOBScf
vyozYlQBjT/sVQM5k6tWbiAagNwQLPSZbzSwNTrf0k4yakU051hf80TS3qTdmkCF
MDWxMlebESwcDoQOVyq7pLNyYN3Y5WwiBlaZZLAD71uztfAPMJQ+DWLq/nhLuE5v
xMTeTBtlN5qy2/LPjrqMW5JoMWzOJO3IZdP8Vo5d4LvC29Cy7UOgEwYsBJ0KDo/Z
Uz8uWucsPevJ5fEmpa6EkWlSm7adky2pMCPdxfdMT+RjYBKShr8NSsXe4ukbXakt
mHYJcqCbKlUy3J1/U1g07JUtv6jKEjUPUdKPqy0iXTglmdeOs5OWfffX9TITNxdQ
31RZjCVwIGiuszfdQ572qGPZw49SitoTMxJshMJEY7QLZm9vQGJhci5jb22JARwE
EAECAAYFAmIqZEEACgkQXAIYP4CH4zaoNQf/cSsNjUq84Frd1l96o3stxw/k7DGa
x1SvQ6T4zTlTndWZirnPv3OncrtzjpXjCFbymxBKhFKgrTA5hgT/Gm0z/BQWEGqf
GtiIcTnoPIJfrxGXYqqaMVMzFUIha3vfqf4DqmOcZIebUmOp926uWKnp7AMFt2Qs
1uhGIq5gxf44zKp3dKfLTU0GaqsGx8F0Hfc8uguXhKB8OdkhmwEwZMdFvTg4y8px
akA1wLgb4M2/j876jL16SacsmrOfD/f9OygPCUJWOPf83frmcJHrJJGeJ2G+V4Nv
yxxawXzHgc4q6+YvBjgCn3LWojXqh/07nRLiRNfFNX5DKKSa7AUE4ET4Mg==
=YcmB
-----END PGP PRIVATE KEY BLOCK-----
`,
			Passphrase:    "not-real-passphrase",
			ExpectedError: errors.StructuralError("private key checksum failure"),
		},
		{
			SigningKey: `
-----BEGIN PGP PRIVATE KEY BLOCK-----
woa!
-----END PGP PRIVATE KEY BLOCK-----
`,
			Passphrase:    "abcd123",
			ExpectedError: errors.InvalidArgumentError("no armored data found"),
		},
	}
)

func TestParseMessage(t *testing.T) {

	for id, data := range dataset {
		gpg := GPGSpec{SigningKey: data.SigningKey, Passphrase: data.Passphrase, Enabled: true}

		_, err := GetCommitSignKey(gpg.SigningKey, gpg.Passphrase)

		if err != nil {
			if strings.Compare(err.Error(), data.ExpectedError.Error()) != 0 {
				t.Errorf("Wrong sign %d err:\n\tExpected:\t\t%v\n\tGot:\t\t%v\n", id, data.ExpectedError, err)
			}
		} else if err != nil {
			t.Errorf("Unexpected error %q for sign #%d", err, id)
		}
	}
}
