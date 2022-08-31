#!/bin/sh


MOVIE_NAME=$1
TEMP_URL=$2

expected_content_type="image/jpeg"
content_type=$(curl -sL -I -o /dev/null -w '%{content_type}' ${TEMP_URL})

check_url=$(echo ${TEMP_URL} | grep "performed_by")
if [ ! -z ${check_url} ]; then
	TEMP_URL=$(echo ${TEMP_URL} | sed 's/\([0-9]\+\)\/download.*/\1/')
fi

MOVIE_NAME=$(echo ${MOVIE_NAME} | sed 's/^Missing poster for movie //')

if [ "${content_type}" != "${expected_content_type}" ]; then
	echo "Invalid content type returned for URL ${TEMP_URL}: \`${content_type}\`"
	exit 1
fi

temp_url=${TEMP_URL} metadata_path=".metadata.[\"${MOVIE_NAME}\"].url_poster" yq -i 'eval(strenv(metadata_path)) = strenv(temp_url)' config/movies.yml
