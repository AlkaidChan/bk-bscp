<template>
  <section class="publish-version" v-if="versionData.status.publish_status === 'not_released'">
    <bk-button
      v-if="
        !approveData?.status ||
        ![APPROVE_STATUS.pending_publish, APPROVE_STATUS.already_publish].includes(approveData.status as APPROVE_STATUS)
      "
      v-cursor="{ active: !props.hasPerm }"
      theme="primary"
      :class="['trigger-button', { 'bk-button-with-no-perm': !props.hasPerm }]"
      :disabled="props.permCheckLoading || approveData?.status === APPROVE_STATUS.pending_approval"
      @click="handleBtnClick">
      {{ t('上线版本') }}
    </bk-button>
    <bk-button
      v-if="approveData?.status === APPROVE_STATUS.pending_publish && approveData.type === ONLINE_TYPE.scheduled"
      v-cursor="{ active: !props.hasPerm }"
      v-bk-tooltips="{
        disabled: approveData.type !== ONLINE_TYPE.scheduled,
        content: `定时上线：${convertTime(approveData.time, 'local')}`,
        placement: 'bottom-end',
      }"
      theme="primary"
      :class="['trigger-button', { 'bk-button-with-no-perm': !props.hasPerm }]"
      :disabled="approveData.type === ONLINE_TYPE.scheduled"
      @click="handlePublishClick">
      <!-- 审批通过时间在定时上线时间之后，后端自动转为手动上线 -->
      {{ t('等待定时上线') }}
    </bk-button>
    <bk-button
      v-if="approveData?.status === APPROVE_STATUS.pending_publish && approveData.type === ONLINE_TYPE.manually"
      v-cursor="{ active: !props.hasPerm }"
      v-bk-tooltips="{
        disabled: props.creator === userInfo.username,
        content: $t('无确认上线权限文案', { creator: props.creator }),
        placement: 'bottom-end',
      }"
      theme="primary"
      :class="['trigger-button', { 'bk-button-with-no-perm': !props.hasPerm }]"
      :disabled="props.creator !== userInfo.username"
      @click="handlePublishClick">
      <!-- 审批通过时间在定时上线时间之后，后端自动转为手动上线 -->
      {{ t('确认上线') }}
    </bk-button>
    <Teleport to="body">
      <VersionLayout v-if="isSelectGroupPanelOpen">
        <template #header>
          <section class="header-wrapper">
            <span class="header-name" @click="handlePanelClose">
              <ArrowsLeft class="arrow-left" />
              <span class="service-name">{{ appData.spec.name }}</span>
            </span>
            <AngleRight class="arrow-right" />
            {{ t('上线版本') }}：{{ versionData.spec.name }}
          </section>
        </template>
        <select-group
          :loading="versionListLoading || groupListLoading"
          :group-list="treeNodeGroups"
          :version-list="versionList"
          :release-type="releaseType"
          :groups="groups"
          @open-preview-version-diff="openPreviewVersionDiff"
          @release-type-change="releaseType = $event"
          @change="groups = $event" />
        <template #footer>
          <section class="actions-wrapper">
            <bk-button
              v-bk-tooltips="{ content: t('请选择分组实例'), disabled: groups.length > 0 }"
              class="publish-btn"
              theme="primary"
              :disabled="groups.length === 0"
              @click="handlePublishOrOpenDiff">
              {{ diffableVersionList.length ? t('对比并上线') : t('上线版本') }}
            </bk-button>
            <bk-button @click="handlePanelClose">{{ t('取消') }}</bk-button>
          </section>
        </template>
      </VersionLayout>
    </Teleport>
    <ConfirmDialog
      v-model:show="isConfirmDialogShow"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      :group-list="treeNodeGroups"
      :version-list="versionList"
      :release-type="releaseType"
      :groups="groups"
      :second-confirm="isSecondConfirm"
      :memo="approveData.memo"
      @confirm="handleConfirm"
      @second-confirm="handleSecondConfirm" />
    <PublishVersionDiff
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      :show="isDiffSliderShow"
      :current-version="versionData"
      :base-version-id="baseVersionId"
      :version-list="diffableVersionList"
      :current-version-groups="groupsPendingtoPublish"
      @publish="handleOpenPublishDialog"
      @close="isDiffSliderShow = false" />
    <dialog-warn v-model:show="warnDialogShow" :dialog-data="warnDialogData" @confirm="dialogConfirm" />
  </section>
</template>
<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter, useRoute } from 'vue-router';
  import { ArrowsLeft, AngleRight } from 'bkui-vue/lib/icon';
  import { InfoBox } from 'bkui-vue';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../../../store/global';
  import { IGroupToPublish, IGroupItemInService } from '../../../../../../../types/group';
  import useServiceStore from '../../../../../../store/service';
  import useConfigStore from '../../../../../../store/config';
  import useUserStore from '../../../../../../store/user';
  import { getConfigVersionList, versionStatusCheck } from '../../../../../../api/config';
  import { approve } from '../../../../../../api/record';
  import { getServiceGroupList } from '../../../../../../api/group';
  import { IConfigVersion, IReleasedGroup } from '../../../../../../../types/config';
  import { IPublishData } from '../../../../../../../types/service';
  import VersionLayout from '../../config/components/version-layout.vue';
  import ConfirmDialog from './confirm-dialog.vue';
  import SelectGroup from './select-group/index.vue';
  import PublishVersionDiff from '../publish-version-diff.vue';
  import { ONLINE_TYPE, APPROVE_STATUS } from '../../../../../../constants/record';
  import DialogWarn from '../dialog-publish-warn.vue';
  import { convertTime } from '../../../../../../utils';
  import dayjs from 'dayjs';
  import BkMessage from 'bkui-vue/lib/message';

  const { permissionQuery, showApplyPermDialog } = storeToRefs(useGlobalStore());
  const serviceStore = useServiceStore();
  const versionStore = useConfigStore();
  const { appData } = storeToRefs(serviceStore);
  const { versionData, publishedVersionId } = storeToRefs(versionStore);
  const { userInfo } = storeToRefs(useUserStore());
  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    permCheckLoading: boolean;
    hasPerm: boolean;
    approveData: {
      status: string;
      time: string;
      type: string;
      memo: string;
      groupIds: number[];
    };
    creator: string;
  }>();

  const emit = defineEmits(['confirm']);

  const router = useRouter();
  const route = useRoute();
  const versionList = ref<IConfigVersion[]>([]);
  const versionListLoading = ref(true);
  const groupList = ref<IGroupItemInService[]>([]);
  const groupListLoading = ref(true);
  const treeNodeGroups = ref<IGroupToPublish[]>([]);
  const isSelectGroupPanelOpen = ref(false);
  const isDiffSliderShow = ref(false);
  const isConfirmDialogShow = ref(false);
  const releaseType = ref('select');
  const groups = ref<IGroupToPublish[]>([]);
  const baseVersionId = ref(0);
  const warnDialogShow = ref(false);
  const warnDialogData = ref<IPublishData>({
    final_approval_time: '', // 最后更新时间
    version_name: '', // 最后上线的版本名称
    release_id: 0, // 最后上线的版本id
  });

  const isSecondConfirm = ref(false); // 是否为二次确认

  const permissionQueryResource = computed(() => [
    {
      biz_id: props.bkBizId,
      basic: {
        type: 'app',
        action: 'publish',
        resource_id: props.appId,
      },
    },
  ]);

  // 待上线分组实例
  const groupsPendingtoPublish = computed(() => {
    const list: IReleasedGroup[] = [];
    groups.value.forEach((item) => {
      const group = groupList.value.find((g) => g.group_id === item.id);
      if (group) {
        const { group_id, group_name, new_selector, old_selector, edited } = group;
        list.push({
          id: group_id,
          name: group_name,
          new_selector,
          old_selector,
          edited,
          uid: '',
          mode: '',
        });
      }
    });
    return list;
  });

  // 包含分组变更的版本，用来对比线上版本
  const diffableVersionList = computed(() => {
    const list = [] as IConfigVersion[];
    versionList.value.forEach((version) => {
      if (version.id === versionData.value.id) return; // 忽略当前上线版本
      version.status.released_groups.some((item) => {
        // 其他版本包含默认分组，且当前选中分组未上线
        if (item.id === 0) {
          return groups.value.some((g) => {
            // 选中未上线分组或默认分组
            if (g.release_id === 0 || g.id === 0) {
              list.push(version);
              return true;
            }
            return false;
          });
        }
        // 其他版本包含的分组在当前已选中的分组中
        if (groups.value.some((g) => g.id === item.id)) {
          list.push(version);
          return true;
        }
        return false;
      });
    });
    return list;
  });

  watch(
    () => props.approveData.status,
    (newV, oldV) => {
      if (newV === APPROVE_STATUS.pending_publish) {
        isSecondConfirm.value = true;
        handlePendingPublish();
      } else {
        isSecondConfirm.value = false;
      }
      if (newV !== oldV) {
        handlePanelClose();
      }
    },
  );

  // 获取所有分组，并组装tree组件节点需要的数据
  const getAllGroupData = async () => {
    groupListLoading.value = true;
    const res = await getServiceGroupList(props.bkBizId, appData.value.id as number);
    groupList.value = res.details;
    treeNodeGroups.value = res.details.map((group: IGroupItemInService) => {
      const { group_id, group_name, release_id, release_name } = group;
      const selector = group.new_selector;
      const rules = selector.labels_and || selector.labels_or || [];
      return { id: group_id, name: group_name, release_id, release_name, rules };
    });
    groupListLoading.value = false;
  };

  /**
   * @description 直接上线或先对比再上线
   * 所有分组都为首次上线，则直接上线，反之先对比再上线
   */
  const handlePublishOrOpenDiff = () => {
    if (diffableVersionList.value.length) {
      baseVersionId.value = diffableVersionList.value[0].id;
      isDiffSliderShow.value = true;
      return;
    }
    handleOpenPublishDialog();
  };

  // 获取所有已上线版本（已上线或灰度中）
  const getVersionList = async () => {
    try {
      versionListLoading.value = true;
      const res = await getConfigVersionList(props.bkBizId, props.appId, { start: 0, all: true });
      versionList.value = res.data.details.filter((item: IConfigVersion) => {
        const { id, status } = item;
        return id !== versionData.value.id && status.publish_status !== 'not_released';
      });
      versionListLoading.value = false;
    } catch (e) {
      console.error(e);
    }
  };

  // 风险提示弹窗→继续上线
  const dialogConfirm = (isContinue: boolean) => {
    warnDialogShow.value = false;
    // 继续上线
    if (isContinue) {
      continuePublish();
    }
  };

  // 上线
  const handleBtnClick = async () => {
    const checkResult = await checkVersionStatus();
    if (checkResult) {
      continuePublish();
    }
  };

  const continuePublish = () => {
    getVersionList();
    getAllGroupData();
    if (props.hasPerm) {
      isSelectGroupPanelOpen.value = true;
    } else {
      permissionQuery.value = { resources: permissionQueryResource.value };
      showApplyPermDialog.value = true;
    }
  };

  const handlePublishClick = async () => {
    isSecondConfirm.value = true;
    isConfirmDialogShow.value = true;
  };

  // 确定上线按钮
  const handleSecondConfirm = async () => {
    const { bkBizId: biz_id, appId: app_id } = props;
    // 上线后查询当前版本状态
    const resp = await approve(biz_id, app_id, versionData.value.id, {
      publish_status: APPROVE_STATUS.already_publish,
    });
    // 这里有两种情况且不会同时出现：
    // 1. itsm已经审批了，但我们产品页面还没有刷新
    // 2. itsm已经撤销了，但我们产品页面还没有刷新
    // 如果存在以上两种情况之一，提示使用message，否则继续后面流程
    const { message } = resp;
    if (message) {
      // 不再走上线流程
      publishedVersionId.value = versionData.value.id;
      emit('confirm', Number(route.params.versionId));
      BkMessage({
        theme: 'primary',
        message,
      });
    } else {
      // 继续上线流程
      handleConfirm(resp.have_pull, false);
    }
  };

  const handleOpenPublishDialog = () => {
    isConfirmDialogShow.value = true;
  };

  // 选择分组面板上线预览版本对比
  const openPreviewVersionDiff = (id: number) => {
    baseVersionId.value = id;
    isDiffSliderShow.value = true;
  };

  // 版本上线文案
  const publishTitle = (isApprove: boolean, type: string, time: string) => {
    switch (type) {
      case ONLINE_TYPE.manually:
        return t('手动上线文案');
      case ONLINE_TYPE.automatically:
        // return t('待审批通过后，调整分组将自动上线');
        return t('审批通过后上线文案');
      case ONLINE_TYPE.scheduled:
        return isApprove ? t('需审批-定时上线文案', { time }) : t('定时上线文案', { time });
      default:
        return t('版本已上线');
    }
  };

  // 版本上线成功/提交成功
  const handleConfirm = (havePull: boolean, isApprove: boolean, publishType = '', publishTime = '') => {
    isDiffSliderShow.value = false;
    publishedVersionId.value = versionData.value.id;
    handlePanelClose();
    emit('confirm', Number(route.params.versionId));
    if (havePull || (!havePull && isApprove)) {
      InfoBox({
        infoType: 'success',
        'ext-cls': 'info-box-style',
        title: publishTitle(isApprove, publishType, publishTime),
        dialogType: 'confirm',
      });
    } else {
      InfoBox({
        infoType: 'success',
        title: publishTitle(isApprove, publishType, publishTime),
        'ext-cls': 'info-box-style',
        confirmText: t('配置客户端'),
        cancelText: t('稍后再说'),
        onConfirm: () => {
          const routeData = router.resolve({
            name: 'configuration-example',
            params: { spaceId: props.bkBizId, appId: props.appId },
          });
          window.open(routeData.href, '_blank');
        },
      });
    }
  };

  const handlePanelClose = () => {
    releaseType.value = 'select';
    isSelectGroupPanelOpen.value = false;
    groups.value = [];
  };

  // 检查是否有正在上线的版本 或 2小时内有无其他版本上线
  const checkVersionStatus = async () => {
    const resp = await versionStatusCheck(props.bkBizId, props.appId);
    const { data } = resp;
    warnDialogData.value = data;
    if (data?.is_publishing) {
      // 当前服务有其他版本上线，不允许当前版本上线
      warnDialogShow.value = true;
      return false;
    }
    if (data?.publish_record.length) {
      // 最近上线的版本时间与当前系统时间在2小时内，风险提示弹窗
      const time = data.publish_record[0].final_approval_time;
      const publishTime = dayjs(convertTime(time, 'local'));
      const currentTime = dayjs();
      if (publishTime.add(2, 'hour').isAfter(currentTime)) {
        warnDialogShow.value = true;
        return false;
      }
    }
    // 可以继续上线
    return true;
  };

  // 待确认上线
  const handlePendingPublish = async () => {
    await getAllGroupData();
    await getVersionList();
    if (props.approveData.groupIds.length === 0) {
      groups.value = treeNodeGroups.value;
      releaseType.value = 'all';
    } else {
      groups.value = treeNodeGroups.value.filter((item) => props.approveData.groupIds.includes(item.id));
      if (props.approveData.groupIds.includes(0)) {
        releaseType.value = 'all';
      } else {
        releaseType.value = 'select';
      }
    }
  };

  defineExpose({
    openSelectGroupPanel: handleBtnClick,
  });
</script>
<style lang="scss" scoped>
  .publish-version {
    display: flex;
    justify-content: flex-start;
    align-items: center;
  }
  .trigger-button {
    margin-left: 8px;
  }
  .header-wrapper {
    display: flex;
    align-items: center;
    padding: 0 24px;
    height: 100%;
    font-size: 12px;
    line-height: 1;
  }
  .header-name {
    display: flex;
    align-items: center;
    font-size: 12px;
    color: #3a84ff;
    cursor: pointer;
  }
  .arrow-left {
    font-size: 26px;
    color: #3884ff;
  }
  .arrow-right {
    font-size: 24px;
    color: #c4c6cc;
  }
  .actions-wrapper {
    display: flex;
    align-items: center;
    padding: 0 24px;
    height: 100%;
    .publish-btn {
      margin-right: 8px;
    }
    .bk-button {
      min-width: 88px;
    }
  }
  .version-selector {
    display: flex;
    align-items: center;
    height: 100%;
    padding: 0 24px;
    font-size: 12px;
  }
</style>
