#!/usr/bin/env bash
set -euo pipefail

api_gateway_base_url="${API_GATEWAY_BASE_URL:-http://localhost/api/development/v1}"
api_key="${API_KEY:-}"
user_agent="${USER_AGENT:-NotegicCurlExample/1.0}"
id="${AVATAR_ID:-1}"
userPublicId="${USERPUBLICID:-00000000-0000-4000-8000-000000000001}"
stationId="${STATIONID:-00000000-0000-4000-8000-000000000001}"
routineId="${ROUTINEID:-00000000-0000-4000-8000-000000000001}"
routineTagId="${ROUTINETAGID:-00000000-0000-4000-8000-000000000001}"
routineTaskId="${ROUTINETASKID:-00000000-0000-4000-8000-000000000001}"
rootShelfId="${ROOTSHELFID:-00000000-0000-4000-8000-000000000001}"
subShelfId="${SUBSHELFID:-00000000-0000-4000-8000-000000000001}"
prevSubShelfId="${PREVSUBSHELFID:-00000000-0000-4000-8000-000000000001}"
parentSubShelfId="${PARENTSUBSHELFID:-00000000-0000-4000-8000-000000000001}"
materialId="${MATERIALID:-00000000-0000-4000-8000-000000000001}"
blockPackId="${BLOCKPACKID:-00000000-0000-4000-8000-000000000001}"
blockId="${BLOCKID:-00000000-0000-4000-8000-000000000001}"
itemId="${ITEMID:-00000000-0000-4000-8000-000000000001}"

# Run individual functions deliberately. DELETE/reset functions are not invoked automatically.

deleteMyBlockPacksByIds() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"blockPackIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/block-packs/batch"
}

createBlockPacks() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"createdBlockPacks":[{"headerBackgroundURL":"https://example.com","icon":"😀","id":"00000000-0000-4000-8000-000000000001","name":"example","parentSubShelfId":"00000000-0000-4000-8000-000000000001"}]}' \
    "$api_gateway_base_url/block-packs/batch"
}

updateMyBlockPacksByIds() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"updatedBlockPacks":[{"blockPackId":"00000000-0000-4000-8000-000000000001","setNull":{},"values":{"headerBackgroundURL":"https://example.com","icon":"😀","name":"example"}}]}' \
    "$api_gateway_base_url/block-packs/batch"
}

moveMyBlockPacksByParentSubShelfIds() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"movedBlockPacks":[{"blockPackIds":["00000000-0000-4000-8000-000000000001"],"destinationParentSubShelfId":"00000000-0000-4000-8000-000000000001"}]}' \
    "$api_gateway_base_url/block-packs/batch/position"
}

restoreMyBlockPacksByIds() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"blockPackIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/block-packs/batch/restore"
}

moveMyBlockPacksByParentSubShelfId() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"blockPackIds":["00000000-0000-4000-8000-000000000001"],"destinationParentSubShelfId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/block-packs/position"
}

getAllMyBlockPacksByRootShelfId() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/block-packs/root-shelf/${rootShelfId}?areDeleted=true"
}

getMyBlockPacksByParentSubShelfId() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/block-packs/sub-shelf/${parentSubShelfId}?areDeleted=true"
}

createBlockPack() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"headerBackgroundURL":"https://example.com","icon":"😀","id":"00000000-0000-4000-8000-000000000001","name":"example","parentSubShelfId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/block-packs/sub-shelf/${parentSubShelfId}"
}

deleteMyBlockPackById() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/block-packs/${blockPackId}"
}

getMyBlockPackById() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/block-packs/${blockPackId}?isDeleted=true"
}

updateMyBlockPackById() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"setNull":{},"values":{"headerBackgroundURL":"https://example.com","icon":"😀","name":"example"}}' \
    "$api_gateway_base_url/block-packs/${blockPackId}"
}

getMyBlockPackAndItsParentById() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/block-packs/${blockPackId}/parent"
}

moveMyBlockPackByParentSubShelfId() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"blockPackId":"00000000-0000-4000-8000-000000000001","destinationParentSubShelfId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/block-packs/${blockPackId}/position"
}

restoreMyBlockPackById() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/block-packs/${blockPackId}/restore"
}

getMyBlocksByIds() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/blocks/batch?blockIds=00000000-0000-4000-8000-000000000001"
}

getMyBlocksByBlockPackId() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/blocks/block-pack/${blockPackId}"
}

getMyBlockById() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/blocks/${blockId}"
}

deleteMyMaterialsByIds() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"materialIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/materials/batch"
}

moveMyMaterialsByIds() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"destinationParentSubShelfId":"00000000-0000-4000-8000-000000000001","materialIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/materials/batch/parent"
}

restoreMyMaterialsByIds() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"materialIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/materials/batch/restore"
}

getAllMyMaterialsByRootShelfId() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/materials/root-shelf/${rootShelfId}?areDeleted=true"
}

getMyMaterialsByParentSubShelfId() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/materials/sub-shelf/${parentSubShelfId}?areDeleted=true"
}

createMyMaterial() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"name":"example","parentSubShelfId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/materials/sub-shelf/${parentSubShelfId}"
}

deleteMyMaterialById() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/materials/${materialId}"
}

getMyMaterialById() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/materials/${materialId}?isDeleted=true"
}

updateMyMaterialById() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"setNull":{},"values":{"name":"example"}}' \
    "$api_gateway_base_url/materials/${materialId}"
}

saveMyMaterialById() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"contentFile":[1]}' \
    "$api_gateway_base_url/materials/${materialId}/content"
}

getMyMaterialAndItsParentById() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/materials/${materialId}/parent"
}

moveMyMaterialById() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"destinationParentSubShelfId":"00000000-0000-4000-8000-000000000001","materialId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/materials/${materialId}/parent"
}

restoreMyMaterialById() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/materials/${materialId}/restore"
}

createRootShelf() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"id":"00000000-0000-4000-8000-000000000001","name":"example"}' \
    "$api_gateway_base_url/root-shelves"
}

deleteMyRootShelvesByIds() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"rootShelfIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/root-shelves/batch"
}

createRootShelves() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"insertedRootShelves":[{"id":"00000000-0000-4000-8000-000000000001","name":"example"}]}' \
    "$api_gateway_base_url/root-shelves/batch"
}

updateMyRootShelvesByIds() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"updatedRootShelves":[{"rootShelfId":"00000000-0000-4000-8000-000000000001","setNull":{},"values":{"name":"example"}}]}' \
    "$api_gateway_base_url/root-shelves/batch"
}

restoreMyRootShelvesByIds() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"rootShelfIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/root-shelves/batch/restore"
}

leaveMyRootShelves() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"rootShelves":[{"rootShelfId":"00000000-0000-4000-8000-000000000001"}]}' \
    "$api_gateway_base_url/root-shelves/memberships/me"
}

deleteMyRootShelfById() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"rootShelfId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/root-shelves/${rootShelfId}"
}

getMyRootShelfById() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/root-shelves/${rootShelfId}?isDeleted=true"
}

updateMyRootShelfById() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"setNull":{},"values":{"name":"example"}}' \
    "$api_gateway_base_url/root-shelves/${rootShelfId}"
}

leaveMyRootShelf() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/root-shelves/${rootShelfId}/memberships/me"
}

transferMyRootShelfOwnership() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"targetUserPublicId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/root-shelves/${rootShelfId}/ownership"
}

deleteMyRootShelfPermissions() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"userPublicIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/root-shelves/${rootShelfId}/permissions"
}

upsertMyRootShelfPermissions() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"permissions":[{"permission":"Read","userPublicId":"00000000-0000-4000-8000-000000000001"}]}' \
    "$api_gateway_base_url/root-shelves/${rootShelfId}/permissions"
}

deleteMyRootShelfPermission() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/root-shelves/${rootShelfId}/permissions/${userPublicId}"
}

getMyRootShelfPermission() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/root-shelves/${rootShelfId}/permissions/${userPublicId}"
}

updateMyRootShelfPermission() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/root-shelves/${rootShelfId}/permissions/${userPublicId}"
}

createMyRootShelfPermission() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"permission":"Read"}' \
    "$api_gateway_base_url/root-shelves/${rootShelfId}/permissions/${userPublicId}"
}

upsertMyRootShelfPermission() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/root-shelves/${rootShelfId}/permissions/${userPublicId}"
}

restoreMyRootShelfById() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"rootShelfId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/root-shelves/${rootShelfId}/restore"
}

getAllMyRoutineTags() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routine-tags?areDeleted=true"
}

createRoutineTag() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"color":"example","icon":"example","id":"00000000-0000-4000-8000-000000000001","name":"example"}' \
    "$api_gateway_base_url/routine-tags"
}

createRoutineTags() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"createdRoutineTags":[{"color":"example","icon":"example","id":"00000000-0000-4000-8000-000000000001","name":"example"}]}' \
    "$api_gateway_base_url/routine-tags/batch"
}

updateMyRoutineTagsByIds() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"updatedRoutineTags":[{"routineTagId":"00000000-0000-4000-8000-000000000001","setNull":{},"values":{"color":"example","icon":"example","name":"example"}}]}' \
    "$api_gateway_base_url/routine-tags/batch"
}

hardDeleteMyRoutineTagsByIds() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"routineTagIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/routine-tags/batch/permanently"
}

getMyRoutineTagById() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routine-tags/${routineTagId}?isDeleted=true"
}

updateMyRoutineTagById() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"setNull":{},"values":{"color":"example","icon":"example","name":"example"}}' \
    "$api_gateway_base_url/routine-tags/${routineTagId}"
}

hardDeleteMyRoutineTagById() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routine-tags/${routineTagId}/permanently"
}

getAllMyRoutineTasks() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routine-tasks?areDeleted=true"
}

hardDeleteMyRoutineTasksByIds() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"routineTaskIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/routine-tasks/batch/permanently"
}

createRoutineTaskByRoutineId() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"maxAttempts":1,"payload":{},"previousRoutineTaskIds":["00000000-0000-4000-8000-000000000001"],"priority":1,"purpose":"GetSubShelf","routineId":"00000000-0000-4000-8000-000000000001","title":"example"}' \
    "$api_gateway_base_url/routine-tasks/routine/${routineId}"
}

getAllMyRoutineTasksByRoutineIds() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routine-tasks/routines?areDeleted=true&routineIds=00000000-0000-4000-8000-000000000001"
}

visualizeMyRoutineTaskPurposeCount() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routine-tasks/visualizations/purpose-count?permission=Read"
}

getMyRoutineTaskById() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routine-tasks/${routineTaskId}?isDeleted=true"
}

updateMyRoutineTaskById() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"routineTaskId":"00000000-0000-4000-8000-000000000001","setNull":{},"values":{"maxAttempts":1,"payload":{},"previousRoutineTaskIds":["00000000-0000-4000-8000-000000000001"],"priority":1,"purpose":"GetSubShelf","routineId":"00000000-0000-4000-8000-000000000001","title":"example"}}' \
    "$api_gateway_base_url/routine-tasks/${routineTaskId}"
}

hardDeleteMyRoutineTaskById() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"routineTaskId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/routine-tasks/${routineTaskId}/permanently"
}

getAllMyRoutinesByTimeRange() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routines?areDeleted=true&from=2026-01-01T00%3A00%3A00Z&stationIds=00000000-0000-4000-8000-000000000001&to=2026-01-01T00%3A00%3A00Z"
}

deleteMyRoutinesByIds() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"routineIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/routines/batch"
}

createRoutinesByStationIds() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"createdRoutines":[{"description":"example","id":"00000000-0000-4000-8000-000000000001","isPinned":true,"period":"Daily","scheduledEndAt":"2026-01-01T00:00:00Z","scheduledStartAt":"2026-01-01T00:00:00Z","stationId":"00000000-0000-4000-8000-000000000001","status":"Scheduled","timezone":"example","title":"example"}]}' \
    "$api_gateway_base_url/routines/batch"
}

updateMyRoutinesByIds() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"updatedRoutines":[{"routineId":"00000000-0000-4000-8000-000000000001","setNull":{},"values":{"description":"example","isPinned":true,"period":"Daily","scheduledEndAt":"2026-01-01T00:00:00Z","scheduledStartAt":"2026-01-01T00:00:00Z","stationId":"00000000-0000-4000-8000-000000000001","status":"Scheduled","timezone":"example","title":"example"}}]}' \
    "$api_gateway_base_url/routines/batch"
}

hardDeleteMyRoutinesByIds() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routines/batch/permanently"
}

restoreMyRoutinesByIds() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"routineIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/routines/batch/restore"
}

linkRoutineItemsByIds() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"isUnlink":true,"linkedRoutinesAndItems":[{"itemId":"00000000-0000-4000-8000-000000000001","itemType":"BlockPack","routineId":"00000000-0000-4000-8000-000000000001"}]}' \
    "$api_gateway_base_url/routines/items"
}

getMyRoutinesByStationId() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routines/station/${stationId}?areDeleted=true"
}

createRoutineByStationId() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"description":"example","id":"00000000-0000-4000-8000-000000000001","isPinned":true,"period":"Daily","scheduledEndAt":"2026-01-01T00:00:00Z","scheduledStartAt":"2026-01-01T00:00:00Z","stationId":"00000000-0000-4000-8000-000000000001","status":"Scheduled","timezone":"example","title":"example"}' \
    "$api_gateway_base_url/routines/station/${stationId}"
}

linkRoutineTagsByIds() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"isUnlink":true,"linkedRoutinesAndTags":[{"routineId":"00000000-0000-4000-8000-000000000001","routineTagId":"00000000-0000-4000-8000-000000000001"}]}' \
    "$api_gateway_base_url/routines/tags"
}

visualizeMyRoutinePeriodCount() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routines/visualizations/period-count"
}

visualizeMyRoutineScheduledEndAtCount() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routines/visualizations/scheduled-end-at-count"
}

visualizeMyRoutineScheduledStartAtCount() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routines/visualizations/scheduled-start-at-count?permission=Read&queryRangeEndedAt=2026-01-01T00%3A00%3A00Z&queryRangeStartedAt=2026-01-01T00%3A00%3A00Z&timeHourUnit=1"
}

visualizeMyRoutineStatusCount() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routines/visualizations/status-count?permission=Read"
}

deleteMyRoutineById() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"routineId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/routines/${routineId}"
}

getMyRoutineById() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routines/${routineId}?isDeleted=true"
}

updateMyRoutineById() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"routineId":"00000000-0000-4000-8000-000000000001","setNull":{},"values":{"description":"example","isPinned":true,"period":"Daily","scheduledEndAt":"2026-01-01T00:00:00Z","scheduledStartAt":"2026-01-01T00:00:00Z","stationId":"00000000-0000-4000-8000-000000000001","status":"Scheduled","timezone":"example","title":"example"}}' \
    "$api_gateway_base_url/routines/${routineId}"
}

linkRoutineItemById() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"isUnlink":true,"itemId":"00000000-0000-4000-8000-000000000001","itemType":"BlockPack","routineId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/routines/${routineId}/items/${itemId}"
}

hardDeleteMyRoutineById() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/routines/${routineId}/permanently"
}

restoreMyRoutineById() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"routineId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/routines/${routineId}/restore"
}

linkRoutineTagById() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"isUnlink":true,"routineId":"00000000-0000-4000-8000-000000000001","routineTagId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/routines/${routineId}/tags/${routineTagId}"
}

getAllMyStations() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/stations?areDeleted=true"
}

createStation() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"description":"example","headerBackgroundURL":"https://example.com","icon":"example","id":"00000000-0000-4000-8000-000000000001","name":"example"}' \
    "$api_gateway_base_url/stations"
}

deleteMyStationsByIds() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"stationIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/stations/batch"
}

createStations() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"createdStations":[{"description":"example","headerBackgroundURL":"https://example.com","icon":"example","id":"00000000-0000-4000-8000-000000000001","name":"example"}]}' \
    "$api_gateway_base_url/stations/batch"
}

updateMyStationsByIds() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"updatedStations":[{"setNull":{},"stationId":"00000000-0000-4000-8000-000000000001","values":{"description":"example","headerBackgroundURL":"https://example.com","icon":"example","name":"example"}}]}' \
    "$api_gateway_base_url/stations/batch"
}

hardDeleteMyStationsByIds() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"stationIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/stations/batch/permanently"
}

restoreMyStationsByIds() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"stationIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/stations/batch/restore"
}

leaveMyStations() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"stations":[{"stationId":"00000000-0000-4000-8000-000000000001"}]}' \
    "$api_gateway_base_url/stations/memberships/me"
}

visualizeMyTotalCount() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/stations/visualizations/total-count?permission=Read"
}

deleteMyStationById() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"stationId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/stations/${stationId}"
}

getMyStationById() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/stations/${stationId}?isDeleted=true"
}

updateMyStationById() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"setNull":{},"values":{"description":"example","headerBackgroundURL":"https://example.com","icon":"example","name":"example"}}' \
    "$api_gateway_base_url/stations/${stationId}"
}

leaveMyStation() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/stations/${stationId}/memberships/me"
}

transferMyStationOwnership() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"targetUserPublicId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/stations/${stationId}/ownership"
}

hardDeleteMyStationById() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"stationId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/stations/${stationId}/permanently"
}

deleteMyStationPermissions() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"userPublicIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/stations/${stationId}/permissions"
}

upsertMyStationPermissions() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"permissions":[{"permission":"Read","userPublicId":"00000000-0000-4000-8000-000000000001"}]}' \
    "$api_gateway_base_url/stations/${stationId}/permissions"
}

deleteMyStationPermission() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/stations/${stationId}/permissions/${userPublicId}"
}

getMyStationPermission() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/stations/${stationId}/permissions/${userPublicId}"
}

updateMyStationPermission() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/stations/${stationId}/permissions/${userPublicId}"
}

createMyStationPermission() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"permission":"Read"}' \
    "$api_gateway_base_url/stations/${stationId}/permissions/${userPublicId}"
}

upsertMyStationPermission() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/stations/${stationId}/permissions/${userPublicId}"
}

restoreMyStationById() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"stationId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/stations/${stationId}/restore"
}

deleteMySubShelvesByIds() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"subShelfIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/sub-shelves/batch"
}

createSubShelvesByRootShelfIds() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"createdSubShelves":[{"id":"00000000-0000-4000-8000-000000000001","name":"example","prevSubShelfId":"00000000-0000-4000-8000-000000000001","rootShelfId":"00000000-0000-4000-8000-000000000001"}]}' \
    "$api_gateway_base_url/sub-shelves/batch"
}

updateMySubShelvesByIds() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"updatedSubShelves":[{"setNull":{},"subShelfId":"00000000-0000-4000-8000-000000000001","values":{"name":"example"}}]}' \
    "$api_gateway_base_url/sub-shelves/batch"
}

moveMySubShelvesByRootShelfIds() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"movedSubShelves":[{"destinationRootShelfId":"00000000-0000-4000-8000-000000000001","destinationSubShelfId":"00000000-0000-4000-8000-000000000001","sourceRootShelfId":"00000000-0000-4000-8000-000000000001","sourceSubShelfIds":["00000000-0000-4000-8000-000000000001"]}]}' \
    "$api_gateway_base_url/sub-shelves/batch/position"
}

restoreMySubShelvesByIds() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"subShelfIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/sub-shelves/batch/restore"
}

moveMySubShelvesByRootShelfId() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"destinationRootShelfId":"00000000-0000-4000-8000-000000000001","destinationSubShelfId":"00000000-0000-4000-8000-000000000001","sourceRootShelfId":"00000000-0000-4000-8000-000000000001","sourceSubShelfIds":["00000000-0000-4000-8000-000000000001"]}' \
    "$api_gateway_base_url/sub-shelves/position"
}

getMySubShelvesByPrevSubShelfId() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/sub-shelves/prev-sub-shelf/${prevSubShelfId}?areDeleted=true"
}

getMySubShelvesAndItemsByPrevSubShelfId() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/sub-shelves/prev-sub-shelf/${prevSubShelfId}/items?areDeleted=true"
}

getAllMySubShelvesByRootShelfId() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/sub-shelves/root-shelf/${rootShelfId}?areDeleted=true"
}

createSubShelfByRootShelfId() {
  curl --fail-with-body --silent --show-error -X POST \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"id":"00000000-0000-4000-8000-000000000001","name":"example","prevSubShelfId":"00000000-0000-4000-8000-000000000001","rootShelfId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/sub-shelves/root-shelf/${rootShelfId}"
}

deleteMySubShelfById() {
  curl --fail-with-body --silent --show-error -X DELETE \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/sub-shelves/${subShelfId}"
}

getMySubShelfById() {
  curl --fail-with-body --silent --show-error -X GET \
    -H "User-Agent: $user_agent" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/sub-shelves/${subShelfId}?isDeleted=true"
}

updateMySubShelfById() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"setNull":{},"values":{"name":"example"}}' \
    "$api_gateway_base_url/sub-shelves/${subShelfId}"
}

moveMySubShelfByRootShelfId() {
  curl --fail-with-body --silent --show-error -X PUT \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    --data '{"destinationRootShelfId":"00000000-0000-4000-8000-000000000001","destinationSubShelfId":"00000000-0000-4000-8000-000000000001","sourceRootShelfId":"00000000-0000-4000-8000-000000000001","sourceSubShelfId":"00000000-0000-4000-8000-000000000001"}' \
    "$api_gateway_base_url/sub-shelves/${subShelfId}/position"
}

restoreMySubShelfById() {
  curl --fail-with-body --silent --show-error -X PATCH \
    -H "User-Agent: $user_agent" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $api_key" \
    "$api_gateway_base_url/sub-shelves/${subShelfId}/restore"
}
