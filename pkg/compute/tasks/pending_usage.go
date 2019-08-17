// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tasks

import (
	"context"

	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"

	"yunion.io/x/onecloud/pkg/cloudcommon/db/taskman"
	"yunion.io/x/onecloud/pkg/compute/models"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/util/rbacutils"
)

func ClearTaskPendingUsage(ctx context.Context, task taskman.ITask, scope rbacutils.TRbacScope, ownerId mcclient.IIdentityProvider, platform []string) error {
	pendingUsage := models.SQuota{}
	err := task.GetPendingUsage(&pendingUsage)
	if err != nil {
		log.Errorf("GetPendingUsage fail %s", err)
		return errors.Wrap(err, "task.GetPendingUsage")
	}
	err = models.QuotaManager.CancelPendingUsage(ctx, task.GetUserCred(), scope, ownerId, platform, &pendingUsage, &pendingUsage)
	if err != nil {
		return errors.Wrap(err, "models.QuotaManager.CancelPendingUsage")
	}
	err = task.ClearPendingUsage()
	if err != nil {
		return errors.Wrap(err, "task.ClearPendingUsage")
	}
	return nil
}
