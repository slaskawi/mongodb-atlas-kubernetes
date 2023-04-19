#!/bin/sh

set -eou pipefail

export OPERATOR_IMAGE_VERSION="quay.io/mongodb/mongodb-atlas-kubernetes-operator:${VERSION}"
export OPERATOR_PATH="operators/mongodb-atlas-kubernetes"

cd "$OPENSHIFT_REPOSITORY"
git config --global --add safe.directory /github/workspace/"${OPENSHIFT_REPOSITORY}"

echo "Sync fork"
gh auth setup-git
# gh repo set-default "$OPENSHIFT_REPOSITORY"
gh repo sync -b main

echo "Create release folder"
mkdir -p "${OPERATOR_PATH}/${VERSION}"

echo "Copying bundle data"
cd ../../mongodb-atlas-kubernetes || exit 1
cp -r bundle.Dockerfile bundle/manifests bundle/metadata bundle/tests "../${OPENSHIFT_REPOSITORY}/${OPERATOR_PATH}/${VERSION}"

echo "Entering release folder"
cd "../$OPENSHIFT_REPOSITORY/$OPERATOR_PATH"

if [ "${CERTIFIED}" = "false" ]; then
  echo "Applying change for non-certified release"
  sed -i.bak 's/COPY bundle\/manifests/COPY manifests/' "${VERSION}/bundle.Dockerfile"
  sed -i.bak 's/COPY bundle\/metadata/COPY metadata/' "${VERSION}/bundle.Dockerfile"
  sed -i.bak 's/COPY bundle\/tests\/scorecard/COPY tests\/scorecard/' "${VERSION}/bundle.Dockerfile"
  rm "${VERSION}/bundle.Dockerfile.bak"

  yq e -i '.metadata.annotations.containerImage = env(OPERATOR_IMAGE_VERSION)' "${VERSION}"/manifests/mongodb-atlas-kubernetes.clusterserviceversion.yaml
  yq e -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image = env(OPERATOR_IMAGE_VERSION)' "${VERSION}"/manifests/mongodb-atlas-kubernetes.clusterserviceversion.yaml
else
  echo "Applying change for certified release"
  docker pull "${OPERATOR_IMAGE_VERSION}"
  IMAGE_DIGEST=$(docker inspect --format='{{ index .RepoDigests 0}}' "${OPERATOR_IMAGE_VERSION}")
  export IMAGE_DIGEST

  yq e -i '.metadata.annotations.containerImage = env(IMAGE_DIGEST)' "${VERSION}"/manifests/mongodb-atlas-kubernetes.clusterserviceversion.yaml
  yq e -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image = env(IMAGE_DIGEST)' "${VERSION}"/manifests/mongodb-atlas-kubernetes.clusterserviceversion.yaml

  # Add skip range
  value='">=0.8.0"' yq e -i '.spec.skipRange = env(value)' "${VERSION}"/manifests/mongodb-atlas-kubernetes.clusterserviceversion.yaml
fi

echo "Configure git"
git config --global user.email "41898282+github-actions[bot]@users.noreply.github.com"
git config --global user.name "github-actions[bot]"

echo "Add files, commit, and push"
git checkout -b "mongodb-atlas-kubernetes-operator-${VERSION}"
git add "${VERSION}"
git status
git commit -m "MongoDB Atlas Operator ${VERSION}" --signoff
git push origin "mongodb-atlas-kubernetes-operator-${VERSION}"

#          if [ -f "docs/pull_request_template.md" ]; then
#            # open PR
#            gh pr create --title "operator mongodb-atlas-kubernetes (${VERSION})" \
#              --assignee "${ASSIGNEES}" \
#              --body-file docs/pull_request_template.md
#          else
#            gh pr create --title "operator mongodb-atlas-kubernetes (${VERSION})" \
#              --assignee "${ASSIGNEES}" \
#              --body ""
 #         fi
