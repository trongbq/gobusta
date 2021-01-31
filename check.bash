#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

EXIT_CODE=0

err() {
  echo -e "${RED}> Failed!${NC}"
  EXIT_CODE=1
}

cmd() {
  echo "> $@"
  $@ || err
}

check_vet() {
  cmd go vet -all ./...
}

check_staticcheck() {
  cmd staticcheck ./...
}

main() {
  case "$1" in
    "")
      check_vet
      check_staticcheck
      ;;
    *)
      echo "Unknown command"
      exit
  esac
  if [[ $EXIT_CODE != 0 ]]; then
      echo -e "${RED}FAILED!, check errors above${NC}"
  else
      echo -e "${GREEN}SUCCESS!${NC}"
  fi
  exit $EXIT_CODE
}

main $@
