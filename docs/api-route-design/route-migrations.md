# Route Migration Design

## Purpose

This document is the backend-owned migration inventory for the resource-oriented
REST API. It records retired operation-style paths and their canonical
replacements so route registrations, binders, tests, and downstream API
consumers can be migrated consistently.

All REST routes are rooted at `/api/development/v1`. Collections use plural,
kebab-cased paths; item identifiers belong in the URI. `batch` represents a
bulk collection operation, and a named child path represents a state change.

## Applied

### Root shelves

| Previous contract | Canonical contract |
| --- | --- |
| `GET /rootShelf/getMyRootShelfById?rootShelfId={id}` | `GET /root-shelves/{root-shelf-id}` |
| `POST /rootShelf/createRootShelf` | `POST /root-shelves` |
| `POST /rootShelf/createRootShelves` | `POST /root-shelves/batch` |
| `PUT /rootShelf/updateMyRootShelfById` | `PUT /root-shelves/{root-shelf-id}` |
| `PUT /rootShelf/updateMyRootShelvesByIds` | `PUT /root-shelves/batch` |
| `PATCH /rootShelf/restoreMyRootShelfById` | `PATCH /root-shelves/{root-shelf-id}/restore` |
| `PATCH /rootShelf/restoreMyRootShelvesByIds` | `PATCH /root-shelves/batch/restore` |
| `DELETE /rootShelf/deleteMyRootShelfById` | `DELETE /root-shelves/{root-shelf-id}` |
| `DELETE /rootShelf/deleteMyRootShelvesByIds` | `DELETE /root-shelves/batch` |
| `POST /rootShelf/{root-shelf-id}/ownership-transfer` | `POST /root-shelves/{root-shelf-id}/ownership` |
| `DELETE /rootShelf/{root-shelf-id}/leave` | `DELETE /root-shelves/{root-shelf-id}/memberships/me` |
| `DELETE /rootShelf/leave` | `DELETE /root-shelves/memberships/me` |

`GET /rootShelf/searchRecentRootShelves` is removed. Use GraphQL
`searchRootShelves` for root-shelf searching.

### Stations

| Previous contract | Canonical contract |
| --- | --- |
| `GET /station/getMyStationById?stationId={id}` | `GET /stations/{station-id}` |
| `GET /station/getAllMyStations` | `GET /stations` |
| `POST /station/createStation` | `POST /stations` |
| `POST /station/createStations` | `POST /stations/batch` |
| `PUT /station/updateMyStationById` | `PUT /stations/{station-id}` |
| `PUT /station/updateMyStationsByIds` | `PUT /stations/batch` |
| `PATCH /station/restoreMyStationById` | `PATCH /stations/{station-id}/restore` |
| `PATCH /station/restoreMyStationsByIds` | `PATCH /stations/batch/restore` |
| `DELETE /station/deleteMyStationById` | `DELETE /stations/{station-id}` |
| `DELETE /station/deleteMyStationsByIds` | `DELETE /stations/batch` |
| `DELETE /station/hardDeleteMyStationById` | `DELETE /stations/{station-id}/permanently` |
| `DELETE /station/hardDeleteMyStationsByIds` | `DELETE /stations/batch/permanently` |
| `GET /station/visualizeMyTotalCount` | `GET /stations/visualizations/total-count` |
| `POST /station/{station-id}/ownership-transfer` | `POST /stations/{station-id}/ownership` |
| `DELETE /station/{station-id}/leave` | `DELETE /stations/{station-id}/memberships/me` |
| `DELETE /station/leave` | `DELETE /stations/memberships/me` |

### Sub shelves

| Previous contract | Canonical contract |
| --- | --- |
| `GET /subShelf/getMySubShelfById?subShelfId={id}` | `GET /sub-shelves/{sub-shelf-id}` |
| `GET /subShelf/getMySubShelvesByPrevSubShelfId?prevSubShelfId={id}` | `GET /sub-shelves/prev-sub-shelf/{prev-sub-shelf-id}` |
| `GET /subShelf/getAllMySubShelvesByRootShelfId?rootShelfId={id}` | `GET /sub-shelves/root-shelf/{root-shelf-id}` |
| `GET /subShelf/getMySubShelvesAndItemsByPrevSubShelfId?prevSubShelfId={id}` | `GET /sub-shelves/prev-sub-shelf/{prev-sub-shelf-id}/items` |
| `POST /subShelf/createSubShelfByRootShelfId` | `POST /sub-shelves/root-shelf/{root-shelf-id}` |
| `POST /subShelf/createSubShelvesByRootShelfIds` | `POST /sub-shelves/batch` |
| `PUT /subShelf/updateMySubShelfById` | `PUT /sub-shelves/{sub-shelf-id}` |
| `PUT /subShelf/updateMySubShelvesByIds` | `PUT /sub-shelves/batch` |
| `PUT /subShelf/moveMySubShelf` | `PUT /sub-shelves/{sub-shelf-id}/position` |
| `PUT /subShelf/moveMySubShelvesByRootShelfId` | `PUT /sub-shelves/position` |
| `PUT /subShelf/moveMySubShelvesByRootShelfIds` | `PUT /sub-shelves/batch/position` |
| `PATCH /subShelf/restoreMySubShelfById` | `PATCH /sub-shelves/{sub-shelf-id}/restore` |
| `PATCH /subShelf/restoreMySubShelvesByIds` | `PATCH /sub-shelves/batch/restore` |
| `DELETE /subShelf/deleteMySubShelfById` | `DELETE /sub-shelves/{sub-shelf-id}` |
| `DELETE /subShelf/deleteMySubShelvesByIds` | `DELETE /sub-shelves/batch` |

### Block packs

| Previous contract | Canonical contract |
| --- | --- |
| `GET /blockPack/getMyBlockPackById?blockPackId={id}` | `GET /block-packs/{block-pack-id}` |
| `GET /blockPack/getMyBlockPackAndItsParentById?blockPackId={id}` | `GET /block-packs/{block-pack-id}/parent` |
| `GET /blockPack/getMyBlockPacksByParentSubShelfId?parentSubShelfId={id}` | `GET /block-packs/sub-shelf/{parent-sub-shelf-id}` |
| `GET /blockPack/getAllMyBlockPacksByRootShelfId?rootShelfId={id}` | `GET /block-packs/root-shelf/{root-shelf-id}` |
| `POST /blockPack/createBlockPack` | `POST /block-packs/sub-shelf/{parent-sub-shelf-id}` |
| `POST /blockPack/createBlockPacks` | `POST /block-packs/batch` |
| `PUT /blockPack/updateMyBlockPackById` | `PUT /block-packs/{block-pack-id}` |
| `PUT /blockPack/updateMyBlockPacksByIds` | `PUT /block-packs/batch` |
| `PUT /blockPack/moveMyBlockPackById` | `PUT /block-packs/{block-pack-id}/position` |
| `PUT /blockPack/moveMyBlockPacksByParentSubShelfId` | `PUT /block-packs/position` |
| `PUT /blockPack/moveMyBlockPacksByParentSubShelfIds` | `PUT /block-packs/batch/position` |
| `PATCH /blockPack/restoreMyBlockPackById` | `PATCH /block-packs/{block-pack-id}/restore` |
| `PATCH /blockPack/restoreMyBlockPacksByIds` | `PATCH /block-packs/batch/restore` |
| `DELETE /blockPack/deleteMyBlockPackById` | `DELETE /block-packs/{block-pack-id}` |
| `DELETE /blockPack/deleteMyBlockPacksByIds` | `DELETE /block-packs/batch` |

### Materials

| Previous contract | Canonical contract |
| --- | --- |
| `GET /material/getMyMaterialById?materialId={id}` | `GET /materials/{material-id}` |
| `GET /material/getMyMaterialAndItsParentById?materialId={id}` | `GET /materials/{material-id}/parent` |
| `GET /material/getMyMaterialsByParentSubShelfId?parentSubShelfId={id}` | `GET /materials/sub-shelf/{parent-sub-shelf-id}` |
| `GET /material/getAllMyMaterialsByRootShelfId?rootShelfId={id}` | `GET /materials/root-shelf/{root-shelf-id}` |
| `POST /material/createMyMaterial` | `POST /materials/sub-shelf/{parent-sub-shelf-id}` |
| `PUT /material/updateMyMaterialById` | `PUT /materials/{material-id}` |
| `PUT /material/saveMyMaterialById` | `PUT /materials/{material-id}/content` |
| `PUT /material/moveMyMaterialById` | `PUT /materials/{material-id}/parent` |
| `PUT /material/moveMyMaterialsByIds` | `PUT /materials/batch/parent` |
| `PATCH /material/restoreMyMaterialById` | `PATCH /materials/{material-id}/restore` |
| `PATCH /material/restoreMyMaterialsByIds` | `PATCH /materials/batch/restore` |
| `DELETE /material/deleteMyMaterialById` | `DELETE /materials/{material-id}` |
| `DELETE /material/deleteMyMaterialsByIds` | `DELETE /materials/batch` |

### Blocks and routine tags

| Previous contract | Canonical contract |
| --- | --- |
| `GET /block/getMyBlockById?blockId={id}` | `GET /blocks/{block-id}` |
| `GET /block/getMyBlocksByIds` | `GET /blocks/batch` |
| `GET /block/getMyBlocksByBlockPackId?blockPackId={id}` | `GET /blocks/block-pack/{block-pack-id}` |
| `GET /routineTag/getMyRoutineTagById?routineTagId={id}` | `GET /routine-tags/{routine-tag-id}` |
| `GET /routineTag/getAllMyRoutineTags` | `GET /routine-tags` |
| `POST /routineTag/createRoutineTag` | `POST /routine-tags` |
| `POST /routineTag/createRoutineTags` | `POST /routine-tags/batch` |
| `PUT /routineTag/updateMyRoutineTagById` | `PUT /routine-tags/{routine-tag-id}` |
| `PUT /routineTag/updateMyRoutineTagsByIds` | `PUT /routine-tags/batch` |
| `DELETE /routineTag/hardDeleteMyRoutineTagById` | `DELETE /routine-tags/{routine-tag-id}/permanently` |
| `DELETE /routineTag/hardDeleteMyRoutineTagsByIds` | `DELETE /routine-tags/batch/permanently` |

### User resources

| Previous contract | Canonical contract |
| --- | --- |
| `GET,PUT /userAccount/getMyAccount,updateMyAccount` | `GET,PUT /me/account` |
| `PUT /userAccount/bindGoogleAccount` | `PUT /me/account/google` |
| `PUT /userAccount/unbindGoogleAccount` | `DELETE /me/account/google` |
| `GET,PUT /userInfo/getMyInfo,updateMyInfo` | `GET,PUT /me/info` |
| `GET /userSetting/getMySetting` | `GET /me/settings` |
| `GET /user/getUserData` | `GET /users/data` |
| `GET,PUT /user/getMe,updateMe` | `GET,PUT /users/me` |

### Routines and durable jobs

| Previous contract | Canonical contract |
| --- | --- |
| `GET /routine/getMyRoutineById?routineId={id}` | `GET /routines/{routine-id}` |
| `GET /routine/getMyRoutinesByStationId?stationId={id}` | `GET /routines/station/{station-id}` |
| `GET /routine/getAllMyRoutinesByTimeRange` | `GET /routines` |
| `POST /routine/createRoutineByStationId` | `POST /routines/station/{station-id}` |
| `POST /routine/createRoutinesByStationIds` | `POST /routines/batch` |
| `PUT /routine/updateMyRoutineById` | `PUT /routines/{routine-id}` |
| `PUT /routine/updateMyRoutinesByIds` | `PUT /routines/batch` |
| `POST /routine/linkRoutineTagById` | `POST /routines/{routine-id}/tags/{routine-tag-id}` |
| `POST /routine/linkRoutineTagsByIds` | `POST /routines/tags` |
| `POST /routine/linkRoutineItemById` | `POST /routines/{routine-id}/items/{item-id}` |
| `POST /routine/linkRoutineItemsByIds` | `POST /routines/items` |
| `PATCH /routine/restoreMyRoutineById` | `PATCH /routines/{routine-id}/restore` |
| `PATCH /routine/restoreMyRoutinesByIds` | `PATCH /routines/batch/restore` |
| `DELETE /routine/deleteMyRoutineById` | `DELETE /routines/{routine-id}` |
| `DELETE /routine/deleteMyRoutinesByIds` | `DELETE /routines/batch` |
| `DELETE /routine/hardDeleteMyRoutineById` | `DELETE /routines/{routine-id}/permanently` |
| `DELETE /routine/hardDeleteMyRoutinesByIds` | `DELETE /routines/batch/permanently` |
| `GET /routine/visualizeMyRoutine*` | `GET /routines/visualizations/*` |
| `GET /routineTask/getMyRoutineTaskById?routineTaskId={id}` | `GET /routine-tasks/{routine-task-id}` |
| `GET /routineTask/getAllMyRoutineTasksByRoutineIds` | `GET /routine-tasks/routines` |
| `GET /routineTask/getAllMyRoutineTasks` | `GET /routine-tasks` |
| `POST /routineTask/createRoutineTaskByRoutineId` | `POST /routine-tasks/routine/{routine-id}` |
| `PUT /routineTask/updateMyRoutineTaskById` | `PUT /routine-tasks/{routine-task-id}` |
| `PUT /routineTask/pauseMyRoutineTaskById` | `PUT /routine-tasks/{routine-task-id}/suspension` |
| `PUT /routineTask/resumeMyRoutineTaskById` | `DELETE /routine-tasks/{routine-task-id}/suspension` |
| `DELETE /routineTask/hardDeleteMyRoutineTaskById` | `DELETE /routine-tasks/{routine-task-id}/permanently` |
| `DELETE /routineTask/hardDeleteMyRoutineTasksByIds` | `DELETE /routine-tasks/batch/permanently` |
| `GET /routineTask/visualizeMyRoutineTask*` | `GET /routine-tasks/visualizations/*` |
| `GET /routineTaskRecord/getAllMyRoutineTaskRecordsByRoutineTaskId?routineTaskId={id}` | `GET /routine-task-records/routine-task/{routine-task-id}` |
| `GET /routineRecord/getAllMyRoutineRecordsByRoutineId?routineId={id}` | Removed; use ClientGateway GraphQL `searchRoutineRecords` |
| `GET /routineTaskRecord/visualizeMyRoutineTaskRecord*` | `GET /routine-task-records/visualizations/*` |

### Realtime

| Previous contract | Canonical contract |
| --- | --- |
| `GET /realtime/blockPacks/{block-pack-id}/participants` | `GET /realtime/block-pack/{block-pack-id}/participants` |
| `POST /realtime/createMyRealtimeConnectionTicket` | `POST /realtime/connection/ticket` |
| `POST /realtime/createMyBlockPackChannelTicket` | `POST /realtime/channel/block-pack/ticket` |

### Development-only storage

| Previous contract | Canonical contract |
| --- | --- |
| `GET /storage/listAllInTerminal` | `GET /storage/all` |

## Backend migration rules

For item mutation routes, the URI identifier is authoritative. Binders map the
identifier into the request DTO and services must not trust a duplicate body
identifier. Permission middleware remains the sole permission-policy source;
route migration must not weaken or redefine the declared policy.
