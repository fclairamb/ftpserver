# Changelog

## [0.16.2](https://github.com/fclairamb/ftpserver/compare/v0.16.1...v0.16.2) (2026-07-06)


### Bug Fixes

* **deps:** update module github.com/fclairamb/afero-s3 to v0.5.0 ([#1728](https://github.com/fclairamb/ftpserver/issues/1728)) ([68dd2d8](https://github.com/fclairamb/ftpserver/commit/68dd2d81102bdee43f1e382594f08f44a0bd0b4e))
* **deps:** update module github.com/fclairamb/ftpserverlib to v0.32.1 ([#1727](https://github.com/fclairamb/ftpserver/issues/1727)) ([a32f02a](https://github.com/fclairamb/ftpserver/commit/a32f02a7708fcce3ce1d73f8c4a9ace54886c4c2))

## [0.16.1](https://github.com/fclairamb/ftpserver/compare/v0.16.0...v0.16.1) (2026-07-01)


### Bug Fixes

* **deps:** update aws-sdk-go-v2 monorepo ([#1726](https://github.com/fclairamb/ftpserver/issues/1726)) ([4cf4f97](https://github.com/fclairamb/ftpserver/commit/4cf4f97b5f5b8b2e0e2467bbe050c41b7bdf81e0))
* **deps:** update module github.com/fclairamb/afero-gdrive to v0.4.0 ([#1723](https://github.com/fclairamb/ftpserver/issues/1723)) ([7d434f5](https://github.com/fclairamb/ftpserver/commit/7d434f5a3c9a87d78f486e06a0032e0c6fb18ee7))
* **deps:** update module google.golang.org/api to v0.287.0 ([#1725](https://github.com/fclairamb/ftpserver/issues/1725)) ([b72f80a](https://github.com/fclairamb/ftpserver/commit/b72f80aede8025d50d0300e2e35671432ac46f71))

## [0.16.0](https://github.com/fclairamb/ftpserver/compare/v0.15.2...v0.16.0) (2026-06-29)


### ⚠ BREAKING CHANGES

* **sftp:** SFTP accesses must now configure host key verification (known_hosts or host_key), or explicitly set insecure_ignore_host_key to "true". Existing configurations relying on the previous unconditional acceptance will fail to connect until updated.

### Features

* **s3:** add basePath support using afero.BasePathFs ([#1636](https://github.com/fclairamb/ftpserver/issues/1636)) ([8e5a218](https://github.com/fclairamb/ftpserver/commit/8e5a218e6732b7a27f6e397f6e534cbb9602cc17))
* **server:** support symlink creation (SITE SYMLINK) ([#1719](https://github.com/fclairamb/ftpserver/issues/1719)) ([4240bc5](https://github.com/fclairamb/ftpserver/commit/4240bc541c89c7805ac5e09154ad6aab3339efbb)), closes [#980](https://github.com/fclairamb/ftpserver/issues/980)


### Bug Fixes

* **ci:** build release binaries again by allowing Go toolchain download ([#1717](https://github.com/fclairamb/ftpserver/issues/1717)) ([723318e](https://github.com/fclairamb/ftpserver/commit/723318ee9c0f67ef27c30cce6098d41464386c7a)), closes [#1658](https://github.com/fclairamb/ftpserver/issues/1658)
* **deps:** migrate keycloak import from gocloak/v13 to v14 ([#1681](https://github.com/fclairamb/ftpserver/issues/1681)) ([438cc28](https://github.com/fclairamb/ftpserver/commit/438cc288ffa0687940605b9b86269d97b217265f)), closes [#1678](https://github.com/fclairamb/ftpserver/issues/1678)
* **deps:** update aws-sdk-go-v2 monorepo ([#1722](https://github.com/fclairamb/ftpserver/issues/1722)) ([2002595](https://github.com/fclairamb/ftpserver/commit/2002595514106f920a5ba52fbff9975fbdf5db5a))
* **deps:** update module cloud.google.com/go/storage to v1.63.0 ([#1715](https://github.com/fclairamb/ftpserver/issues/1715)) ([e07913d](https://github.com/fclairamb/ftpserver/commit/e07913d2e24115f0363a97957121c4a0dd974614))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/s3 to v1.102.1 ([#1693](https://github.com/fclairamb/ftpserver/issues/1693)) ([f05a3f2](https://github.com/fclairamb/ftpserver/commit/f05a3f2af165ab7b4f99d6e48ae25caf1398b4e4))
* **deps:** update module github.com/fclairamb/ftpserverlib to v0.32.0 ([#1692](https://github.com/fclairamb/ftpserver/issues/1692)) ([87fac19](https://github.com/fclairamb/ftpserver/commit/87fac19eb3f3c083463cdd408cd8b94d399d8738))
* **deps:** update module github.com/go-crypt/crypt to v0.14.15 ([#1686](https://github.com/fclairamb/ftpserver/issues/1686)) ([3c02d74](https://github.com/fclairamb/ftpserver/commit/3c02d74b5d2da04e6e10fa9a08df5e741ca4c1e9))
* **deps:** update module github.com/nerzal/gocloak/v13 to v14 ([#1672](https://github.com/fclairamb/ftpserver/issues/1672)) ([30146c5](https://github.com/fclairamb/ftpserver/commit/30146c534be003aa40976705d6e5f8fb8d4cc8cc))
* **deps:** update module golang.org/x/crypto to v0.53.0 ([#1704](https://github.com/fclairamb/ftpserver/issues/1704)) ([1953ede](https://github.com/fclairamb/ftpserver/commit/1953edee5e78f706b26ff24c0e36110e94957257))
* **deps:** update module google.golang.org/api to v0.286.0 ([#1712](https://github.com/fclairamb/ftpserver/issues/1712)) ([0bc270f](https://github.com/fclairamb/ftpserver/commit/0bc270f93a742ecbc5d14e4cc5756dad1fd9e4e5))
* **s3:** normalize basePath and add documentation ([#1679](https://github.com/fclairamb/ftpserver/issues/1679)) ([acddc1b](https://github.com/fclairamb/ftpserver/commit/acddc1bd7adb9bc7b820a010870abd064050c7f1))
* **server:** default to binary transfer type ([#1718](https://github.com/fclairamb/ftpserver/issues/1718)) ([6d30f3d](https://github.com/fclairamb/ftpserver/commit/6d30f3d81f96a45be146b83a2df5f3233f56398b)), closes [#1532](https://github.com/fclairamb/ftpserver/issues/1532)
* **sftp:** verify SSH host key to prevent MITM ([#1716](https://github.com/fclairamb/ftpserver/issues/1716)) ([3b2b3aa](https://github.com/fclairamb/ftpserver/commit/3b2b3aa26568cf9a7bebb57b585154bacb182575))

## [0.15.2](https://github.com/fclairamb/ftpserver/compare/v0.15.1...v0.15.2) (2026-05-16)


### Bug Fixes

* **deps:** update aws-sdk-go-v2 monorepo ([#1608](https://github.com/fclairamb/ftpserver/issues/1608)) ([93724f9](https://github.com/fclairamb/ftpserver/commit/93724f956a9b6af37fecdbf580b361779f69b3d9))
* **deps:** update aws-sdk-go-v2 monorepo ([#1610](https://github.com/fclairamb/ftpserver/issues/1610)) ([b2b590f](https://github.com/fclairamb/ftpserver/commit/b2b590fb758dc9b8369645872e36cb0eecf3b612))
* **deps:** update aws-sdk-go-v2 monorepo ([#1612](https://github.com/fclairamb/ftpserver/issues/1612)) ([f5525c9](https://github.com/fclairamb/ftpserver/commit/f5525c92df5c16694a7a766dc620bd58a2f1c54a))
* **deps:** update aws-sdk-go-v2 monorepo ([#1617](https://github.com/fclairamb/ftpserver/issues/1617)) ([54d7c13](https://github.com/fclairamb/ftpserver/commit/54d7c13614c5fac74187c474001f61a73da0a434))
* **deps:** update aws-sdk-go-v2 monorepo ([#1639](https://github.com/fclairamb/ftpserver/issues/1639)) ([a0efa08](https://github.com/fclairamb/ftpserver/commit/a0efa08e3fd556d1634d6c04125747dc80ce3bfa))
* **deps:** update aws-sdk-go-v2 monorepo ([#1643](https://github.com/fclairamb/ftpserver/issues/1643)) ([80a62de](https://github.com/fclairamb/ftpserver/commit/80a62dea31c964d671f7064ada2682c3661a1795))
* **deps:** update aws-sdk-go-v2 monorepo ([#1649](https://github.com/fclairamb/ftpserver/issues/1649)) ([41aa4b2](https://github.com/fclairamb/ftpserver/commit/41aa4b26bedf3f1ac82388646729f4cb751d3733))
* **deps:** update aws-sdk-go-v2 monorepo ([#1662](https://github.com/fclairamb/ftpserver/issues/1662)) ([5451f66](https://github.com/fclairamb/ftpserver/commit/5451f662b4545f87305a28a8979b45ddf8dae4c8))
* **deps:** update aws-sdk-go-v2 monorepo ([#1665](https://github.com/fclairamb/ftpserver/issues/1665)) ([fec4488](https://github.com/fclairamb/ftpserver/commit/fec4488d81d3b795bc8d7475aa16627f9a0154cd))
* **deps:** update module cloud.google.com/go/storage to v1.59.1 ([#1584](https://github.com/fclairamb/ftpserver/issues/1584)) ([d7aed86](https://github.com/fclairamb/ftpserver/commit/d7aed862074b57d57c8b7e13347e281a200e1722))
* **deps:** update module cloud.google.com/go/storage to v1.59.2 ([#1594](https://github.com/fclairamb/ftpserver/issues/1594)) ([6d85f50](https://github.com/fclairamb/ftpserver/commit/6d85f50a996971740a5b979532b017122a7a40a2))
* **deps:** update module cloud.google.com/go/storage to v1.60.0 ([#1603](https://github.com/fclairamb/ftpserver/issues/1603)) ([f8fd1e6](https://github.com/fclairamb/ftpserver/commit/f8fd1e66b505231f17d9cec6eee34eaec907fd0c))
* **deps:** update module cloud.google.com/go/storage to v1.61.0 ([#1630](https://github.com/fclairamb/ftpserver/issues/1630)) ([a5ef696](https://github.com/fclairamb/ftpserver/commit/a5ef6963ead06d3e76973e1020f4021fd848a701))
* **deps:** update module cloud.google.com/go/storage to v1.61.1 ([#1632](https://github.com/fclairamb/ftpserver/issues/1632)) ([dc4cac9](https://github.com/fclairamb/ftpserver/commit/dc4cac942b7571aa0dcb8c58ee9b966b7dbf22d6))
* **deps:** update module cloud.google.com/go/storage to v1.61.2 ([#1634](https://github.com/fclairamb/ftpserver/issues/1634)) ([8e4fb9a](https://github.com/fclairamb/ftpserver/commit/8e4fb9a55e3f544aa3df4f72c99d6d5eb18cac8c))
* **deps:** update module cloud.google.com/go/storage to v1.61.3 ([#1638](https://github.com/fclairamb/ftpserver/issues/1638)) ([175f599](https://github.com/fclairamb/ftpserver/commit/175f599f816c6cb647b0532a03c95458e710fbf6))
* **deps:** update module cloud.google.com/go/storage to v1.62.0 ([#1650](https://github.com/fclairamb/ftpserver/issues/1650)) ([6a6523a](https://github.com/fclairamb/ftpserver/commit/6a6523a7967495c6c61dbfcdfc5c86df6dde7a0b))
* **deps:** update module cloud.google.com/go/storage to v1.62.1 ([#1657](https://github.com/fclairamb/ftpserver/issues/1657)) ([035f03e](https://github.com/fclairamb/ftpserver/commit/035f03e5707efe85077d22370f1c5f79cc440c03))
* **deps:** update module github.com/aws/aws-sdk-go-v2/config to v1.32.15 ([#1661](https://github.com/fclairamb/ftpserver/issues/1661)) ([6e27e4c](https://github.com/fclairamb/ftpserver/commit/6e27e4cbe73d4ecd451869f91c1e23e881645ae5))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/s3 to v1.100.0 ([#1663](https://github.com/fclairamb/ftpserver/issues/1663)) ([ff0be2c](https://github.com/fclairamb/ftpserver/commit/ff0be2c3329ff558993babec251d227be878b949))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/s3 to v1.101.0 ([#1670](https://github.com/fclairamb/ftpserver/issues/1670)) ([3791214](https://github.com/fclairamb/ftpserver/commit/37912140cc99c8f47a0b7a9fa1a96f3254744593))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/s3 to v1.96.0 ([#1595](https://github.com/fclairamb/ftpserver/issues/1595)) ([0f421fe](https://github.com/fclairamb/ftpserver/commit/0f421fe2ae7a3176107566cd12c05de62998820f))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/s3 to v1.96.2 ([#1616](https://github.com/fclairamb/ftpserver/issues/1616)) ([1971655](https://github.com/fclairamb/ftpserver/commit/1971655fc182fc4c8aa1554d9009b714ff42bdff))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/s3 to v1.96.4 ([#1624](https://github.com/fclairamb/ftpserver/issues/1624)) ([db41034](https://github.com/fclairamb/ftpserver/commit/db41034f08b7e69b9e036d25d6f2db74ccb0e438))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/s3 to v1.97.0 ([#1637](https://github.com/fclairamb/ftpserver/issues/1637)) ([db58c2e](https://github.com/fclairamb/ftpserver/commit/db58c2e1e146895ccc972d5d629f5e1f13354e3f))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/s3 to v1.97.2 ([#1641](https://github.com/fclairamb/ftpserver/issues/1641)) ([a6c7b08](https://github.com/fclairamb/ftpserver/commit/a6c7b0896d94ccafa60998981b5bf93221f3f5d6))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/s3 to v1.98.0 ([#1645](https://github.com/fclairamb/ftpserver/issues/1645)) ([6325e98](https://github.com/fclairamb/ftpserver/commit/6325e98af68b08189073e67c11270665130c5964))
* **deps:** update module github.com/aws/aws-sdk-go-v2/service/s3 to v1.99.0 ([#1652](https://github.com/fclairamb/ftpserver/issues/1652)) ([c98ff7e](https://github.com/fclairamb/ftpserver/commit/c98ff7e5a60025a97c0dd198ee551f8dec4eac9a))
* **deps:** update module github.com/fclairamb/afero-s3 to v0.4.0 ([#1579](https://github.com/fclairamb/ftpserver/issues/1579)) ([f38d547](https://github.com/fclairamb/ftpserver/commit/f38d5479c9765f64040c3e6297e18412f1e2096a))
* **deps:** update module github.com/fclairamb/ftpserverlib to v0.29.0 ([#1580](https://github.com/fclairamb/ftpserver/issues/1580)) ([d161e11](https://github.com/fclairamb/ftpserver/commit/d161e11658f7f1eaa79b1686b7f6a90ed73f41d4))
* **deps:** update module github.com/fclairamb/ftpserverlib to v0.30.0 ([#1591](https://github.com/fclairamb/ftpserver/issues/1591)) ([b66633c](https://github.com/fclairamb/ftpserver/commit/b66633cc9abc665c43eed5a3d5251ab6212cd396))
* **deps:** update module github.com/go-crypt/crypt to v0.4.10 ([#1635](https://github.com/fclairamb/ftpserver/issues/1635)) ([151d45e](https://github.com/fclairamb/ftpserver/commit/151d45e52e0bcc2359623d43caac9698b11378a3))
* **deps:** update module github.com/go-crypt/crypt to v0.4.12 ([#1647](https://github.com/fclairamb/ftpserver/issues/1647)) ([6483def](https://github.com/fclairamb/ftpserver/commit/6483def6be47cac99117538df95f53577495ec9e))
* **deps:** update module github.com/go-crypt/crypt to v0.4.13 ([#1655](https://github.com/fclairamb/ftpserver/issues/1655)) ([8e7c050](https://github.com/fclairamb/ftpserver/commit/8e7c0506ad1055fd3315a65354ddee1355a734ec))
* **deps:** update module github.com/go-crypt/crypt to v0.4.14 ([#1675](https://github.com/fclairamb/ftpserver/issues/1675)) ([c486422](https://github.com/fclairamb/ftpserver/commit/c4864224454ac44d044a8fc57bbadcea4d46909d))
* **deps:** update module github.com/go-crypt/crypt to v0.4.8 ([#1597](https://github.com/fclairamb/ftpserver/issues/1597)) ([d0d3347](https://github.com/fclairamb/ftpserver/commit/d0d33477c907063cd0d7d54593403c142fc9a065))
* **deps:** update module github.com/go-crypt/crypt to v0.4.9 ([#1607](https://github.com/fclairamb/ftpserver/issues/1607)) ([d36bdb3](https://github.com/fclairamb/ftpserver/commit/d36bdb3d9f9459d2f7a062ff9d17c0ec27a34d01))
* **deps:** update module golang.org/x/crypto to v0.47.0 ([#1582](https://github.com/fclairamb/ftpserver/issues/1582)) ([bc97b65](https://github.com/fclairamb/ftpserver/commit/bc97b659882f5ef7c0cbb8261619dcc1c7c92367))
* **deps:** update module golang.org/x/crypto to v0.48.0 ([#1602](https://github.com/fclairamb/ftpserver/issues/1602)) ([380a5f3](https://github.com/fclairamb/ftpserver/commit/380a5f39002b688ff72f2e5741f279863bb1aaf9))
* **deps:** update module golang.org/x/crypto to v0.49.0 ([#1633](https://github.com/fclairamb/ftpserver/issues/1633)) ([d0ac60b](https://github.com/fclairamb/ftpserver/commit/d0ac60b3e570b1a1754630762671e3fc1b2172ce))
* **deps:** update module golang.org/x/crypto to v0.50.0 ([#1654](https://github.com/fclairamb/ftpserver/issues/1654)) ([1259d65](https://github.com/fclairamb/ftpserver/commit/1259d655c8a4fd40c6bc4c7836d13798de8be157))
* **deps:** update module golang.org/x/crypto to v0.51.0 ([#1674](https://github.com/fclairamb/ftpserver/issues/1674)) ([43d5cb9](https://github.com/fclairamb/ftpserver/commit/43d5cb917559e01fe3e5762a4a296731b74cfaba))
* **deps:** update module golang.org/x/oauth2 to v0.35.0 ([#1601](https://github.com/fclairamb/ftpserver/issues/1601)) ([fa2d567](https://github.com/fclairamb/ftpserver/commit/fa2d567cbb0175c5aaa13ff8744725e84e0d9174))
* **deps:** update module golang.org/x/oauth2 to v0.36.0 ([#1628](https://github.com/fclairamb/ftpserver/issues/1628)) ([25302ec](https://github.com/fclairamb/ftpserver/commit/25302ec1466d6dd63d87777ba3ad09cbda9d519a))
* **deps:** update module google.golang.org/api to v0.260.0 ([#1585](https://github.com/fclairamb/ftpserver/issues/1585)) ([64613ab](https://github.com/fclairamb/ftpserver/commit/64613abe91b11019d72467a06e209e00d60385b4))
* **deps:** update module google.golang.org/api to v0.261.0 ([#1588](https://github.com/fclairamb/ftpserver/issues/1588)) ([ea64397](https://github.com/fclairamb/ftpserver/commit/ea6439753614e046c59b64ba8689a74a9445b579))
* **deps:** update module google.golang.org/api to v0.262.0 ([#1590](https://github.com/fclairamb/ftpserver/issues/1590)) ([7ffc890](https://github.com/fclairamb/ftpserver/commit/7ffc89077a6fb8e3a787652ffd43e5cce81c697a))
* **deps:** update module google.golang.org/api to v0.263.0 ([#1592](https://github.com/fclairamb/ftpserver/issues/1592)) ([b27f50e](https://github.com/fclairamb/ftpserver/commit/b27f50e80c76133c3a6fd30599eaf3267961017e))
* **deps:** update module google.golang.org/api to v0.264.0 ([#1596](https://github.com/fclairamb/ftpserver/issues/1596)) ([2146a6e](https://github.com/fclairamb/ftpserver/commit/2146a6e591d7ff51825d57bdd98fe61ea363e8c6))
* **deps:** update module google.golang.org/api to v0.265.0 ([#1598](https://github.com/fclairamb/ftpserver/issues/1598)) ([8235cf5](https://github.com/fclairamb/ftpserver/commit/8235cf5423e4b0c9e7faebf7843acbf3175896dc))
* **deps:** update module google.golang.org/api to v0.266.0 ([#1604](https://github.com/fclairamb/ftpserver/issues/1604)) ([cec1dac](https://github.com/fclairamb/ftpserver/commit/cec1dac883e0465c70ffed8b26fc8b68731d9717))
* **deps:** update module google.golang.org/api to v0.267.0 ([#1609](https://github.com/fclairamb/ftpserver/issues/1609)) ([0c6acbc](https://github.com/fclairamb/ftpserver/commit/0c6acbcb707202282df89588a133c84036563754))
* **deps:** update module google.golang.org/api to v0.268.0 ([#1613](https://github.com/fclairamb/ftpserver/issues/1613)) ([c97b120](https://github.com/fclairamb/ftpserver/commit/c97b1208f41a036f877c884d197d56ec99ad7362))
* **deps:** update module google.golang.org/api to v0.269.0 ([#1614](https://github.com/fclairamb/ftpserver/issues/1614)) ([2101f2e](https://github.com/fclairamb/ftpserver/commit/2101f2ebeaebad79350665f9e783619c93a8282b))
* **deps:** update module google.golang.org/api to v0.270.0 ([#1629](https://github.com/fclairamb/ftpserver/issues/1629)) ([e1fbc1d](https://github.com/fclairamb/ftpserver/commit/e1fbc1dea07fed13873e7fbdbe567d1ac32325c0))
* **deps:** update module google.golang.org/api to v0.271.0 ([#1631](https://github.com/fclairamb/ftpserver/issues/1631)) ([ddeb274](https://github.com/fclairamb/ftpserver/commit/ddeb274b544d2cc4b092519bccf098fe9f344b17))
* **deps:** update module google.golang.org/api to v0.272.0 ([#1640](https://github.com/fclairamb/ftpserver/issues/1640)) ([4aa17b3](https://github.com/fclairamb/ftpserver/commit/4aa17b3edbd6bac4ec235e93a42f0a455fcbc005))
* **deps:** update module google.golang.org/api to v0.273.0 ([#1642](https://github.com/fclairamb/ftpserver/issues/1642)) ([35ab61d](https://github.com/fclairamb/ftpserver/commit/35ab61d3080086c384ce4eb89e25e3120fc0e414))
* **deps:** update module google.golang.org/api to v0.273.1 ([#1646](https://github.com/fclairamb/ftpserver/issues/1646)) ([45bac33](https://github.com/fclairamb/ftpserver/commit/45bac33e689d43f423a40da7b8fd3fcab3a1a7b5))
* **deps:** update module google.golang.org/api to v0.274.0 ([#1648](https://github.com/fclairamb/ftpserver/issues/1648)) ([1031c6b](https://github.com/fclairamb/ftpserver/commit/1031c6bbe746fe06a262561893ed47ee474d1c93))
* **deps:** update module google.golang.org/api to v0.275.0 ([#1651](https://github.com/fclairamb/ftpserver/issues/1651)) ([f752e37](https://github.com/fclairamb/ftpserver/commit/f752e378705caaff14e4c25b8474c3c4ff84e9f2))
* **deps:** update module google.golang.org/api to v0.276.0 ([#1659](https://github.com/fclairamb/ftpserver/issues/1659)) ([6039302](https://github.com/fclairamb/ftpserver/commit/6039302196ca1461c8b08ced57290db2152f20ae))
* **deps:** update module google.golang.org/api to v0.277.0 ([#1666](https://github.com/fclairamb/ftpserver/issues/1666)) ([0961859](https://github.com/fclairamb/ftpserver/commit/0961859470857ad7fcc80a01be1d59981cc035c7))
* **deps:** update module google.golang.org/api to v0.278.0 ([#1669](https://github.com/fclairamb/ftpserver/issues/1669)) ([49faf03](https://github.com/fclairamb/ftpserver/commit/49faf03c5f53f2792b3b73769308bb96788a98dd))
* **deps:** update module google.golang.org/api to v0.279.0 ([#1676](https://github.com/fclairamb/ftpserver/issues/1676)) ([1534db1](https://github.com/fclairamb/ftpserver/commit/1534db118b4f8bca90a65a8fe6f082b45de20287))
