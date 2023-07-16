#!/bin/bash

registry_url="https://arf.tesla.cn/artifactory/api/npm/gfsh-sdk-npm-local/"
user_name="sa-gfsh-sdk-agent"
password=$GFSH_SDK_PASS
token_val=""

if [ -d open-api-ts ]
then
	rm -r open-api-ts
fi

diff_text=`git diff HEAD HEAD~ --name-only`
diff_file_arr=(${diff_text//"\n"/ })

for file in ${diff_file_arr[@]} 
do
	if [[ $file == *".yaml"* && -f $file ]]
	then
		first_line=`sed -n 1p $file`
		if [[ $first_line == *"openapi:"* ]]
		then
			# get api file name ------- start ----------
			file_path_arr=(${file//'.'/ })
			file_path=${file_path_name_arr[0]}
			file_name_arr=(${file//"/"/ })
			file_name=${file_name_arr[${#file_name_arr[*]} - 1]}
			file_name_arr=(${file_name//"."/ })
			file_name=${file_name_arr[0]}
			# get api file name ------- end ----------

			# create dri ------- start ----------
			if [ ! -d open-api-ts ]
			then
				mkdir open-api-ts
			fi

			if [ ! -d "open-api-ts/${file_name}" ]
			then
				mkdir "open-api-ts/${file_name}"
			fi
			# create dri ------- end ----------

			# create package.json ------- start ----------
			if [ ! -f "open-api-ts/${file_name}/package.json" ]
			then
				touch open-api-ts/$file_name/package.json
				echo -e >> open-api-ts/$file_name/package.json
			fi
                  
      npm config set registry $registry_url

			all_versions=`npm view @benjamin/${file_name}-api versions`
			versions=(${all_versions//","/ })
			version=${versions[${#versions[@]}-2]}
			current_version=(${version//"'"/ })

			if [[ $current_version == "" ]]
			then
				new_version="1.0.0"
			else
				version=(${current_version//"."/ })
				version1=${version[0]}
				version2=${version[1]}
				version3=${version[2]}
				version3=$(($version3+1))
				if [ $version3 == 100 ]
				then
					version3=0
					version2=$(($version2+1))
				fi

				if [ $version2 == 100 ]
				then
					version2=0
					version1=$(($version1+1))
				fi

				seperator="."
				new_version="$version1$seperator$version2$seperator$version3"
			fi

			sed -i "1i\{\n\ \ \"name\": \"@benjamin/${file_name}-api\",\n\ \ \"version\": \"${new_version}\",\n\ \ \"description\": \"\",\n\\ \ \"publishConfig\":{\n\ \ \ \ \"registry\":\"${registry_url}\"\n\ \ },\n\ \ \"main\": \"index.ts\",\n\ \ \"dependencies\": {\n\ \ \ \ \"axios\": \"^0.21.1\"\n\ \ },\n\ \ \"author\": \"bjm\",\n\ \ \"license\": \"ISC\"\n}" "open-api-ts/${file_name}/package.json"
			# create package.json ------- end ----------

			# create api file ------- start ----------
			file_suffix="Api"
			api_file_name="$file_name$file_suffix"
			res=`swagger-typescript-api -p $file -t web/open-api-templates -o open-api-ts/$file_name -n $api_file_name`
			# create api file ------- end ----------

			# create index.ts ------- start ----------
			if [ ! -f open-api-ts/$file_name/index.ts ]
			then
				touch open-api-ts/$file_name/index.ts
				echo -e >> open-api-ts/$file_name/index.ts
			fi

			sed -i "1i\import * as ${file_name} from './${file_name}${file_suffix}';" open-api-ts/$file_name/index.ts
			sed -i "$ a\export const ${file_name}Api = ${file_name};" open-api-ts/$file_name/index.ts
			# create index.ts ------- end ----------

			# deploy ------- start ----------
			cd open-api-ts/$file_name
			if [[ $token_val == "" ]]
			then
				token_res=$(curl -s \
					-H "Accept: application/json" \
					-H "Content-Type:application/json" \
					-X PUT --data "{\"name\": \"${user_name}\", \"password\": \"${password}\"}" \
					$registry_url-/user/org.couchdb.user:$user_name)

				if [[ $token_res != *token* ]]
				then
					echo "get arf token fial"
					exit 0     
				fi

				token_res=${token_res//"\n"/}
				token_res=${token_res//" "/}
				token_val=$(echo "${token_res}" | awk -F"[,:}]" '{for(i=1;i<=NF;i++){if($i~/'token'\042/){print $(i+1)}}}' | tr -d '"' | sed -n ${num}p)
			fi
			
			touch .npmrc
			echo -e >> .npmrc
			sed -i "1i\//arf.tesla.cn/artifactory/api/npm/gfsh-sdk-npm-local/:_authToken=${token_val}" .npmrc
			sed -i "1i\registry=https://arf.tesla.cn/artifactory/api/npm/gfsh-sdk-npm-local/" .npmrc
			
			npm publish

			cd ../../
			# deploy ------- start ----------
		fi
	fi
done
