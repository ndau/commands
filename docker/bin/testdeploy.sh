#!/bin/bash
# deploy single devnet node
# get the directory of this file
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

APP_BASE_NAME=sc-node \
SHA=5 \
NODE_NAME=sc-node-0 \
SNAPSHOT="snapshot-devnet-g" \
PERSISTENT_PEERS="ca195d91c051d91c451483b08bdea0059e39fd34@100.24.11.77:30051" \
NODE_IDENTITY="H4sIAAAAAAACA+2UW2/aMBiGc82viHy7asSOnWAkLnKAUY4BRlm4mRISQgIkwTnQtOK/L5Sp25jU3lRok/LcWLI/+7Nkv0/qho7L9n6Y1ldRuPa9ehg57vetW3wOkijkPgKhRML4ZSy5HgWECAdFTKAgYglBToBEkgSOF7gbkCWpxXieY1GUvlX33vp/yjOImZ+f3xs0n0FaxC5ogvTXpzDK1b5btB1ECKTgDuTWLjvXDAeKCu2p/wjtjc/M0Cs+ael0I4/1AGGXjedx0rVG/gKtbEWaTTuT5Xy48pBVdMNjXsiJvg2268UgMO6ZMMmflMep2DAnrRY4nWpcxa1I/8r/y38oX9l3rDRiH2GC9/JPRPSaf0LkMv+yLMMq/zfJf43ngeU4zE0S0OQBhjqVKW5ARdUkoa0iTYIawUiHbYFqVBc1iojWEcHdeWOc2Rd18OdzyomLQPg/DJLZvwvkUvfTImWhYbJO3vcMZ/1VdY3ZPhruH4wHSTxiuoikWb17Lx9IQrTFsQXKvadL31dnvdX4Sl1XnSe6N+jZ6nLnH7Oeao7bB2yaxNm1tdFRM+J1PpJkVujR5gsSqJ8sGgeFQjrsYNq3vGI5p9/YWg+gFcSDxirI2P5p27OV1uWWtUpiFRUV/zo/APQk630ADAAA" \
  "$DIR/deploycontainer.sh"
