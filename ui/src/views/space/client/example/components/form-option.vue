<template>
  <div class="headline">{{ $t('示例参数') }}</div>
  <bk-form class="form-example-wrap" :model="formData" :rules="rules" form-type="vertical" ref="formRef">
    <bk-form-item property="clientKey" required>
      <template #label>
        {{ $t('客户端密钥') }}
        <info
          class="icon-info"
          v-bk-tooltips="{
            content: $t('用于客户端拉取配置时身份验证，下拉列表只会展示关联过此服务且状态为启用的密钥'),
            placement: 'top',
          }" />
      </template>
      <KeySelect
        ref="keySelectorRef"
        :selected-key-data="props.selectedKeyData"
        @current-key="setCredential"
        @selected-key-data="emits('selected-key-data', $event)" />
    </bk-form-item>
    <!-- 配置项/配置文件 -->
    <bk-form-item v-if="props.configShow" property="configName" :required="props.configShow">
      <template #label>
        {{ $t(props.configLabel) }}
        <info
          class="icon-info"
          v-bk-tooltips="{
            content: $t(`请选择一个${props.configLabel}，用于测试下载相应文件`),
            placement: 'top',
          }" />
      </template>
      <config-selector
        ref="configSelectRef"
        :template-name="props.templateName!"
        :service-type="props.serviceType"
        @select-config="formData.configName = $event" />
    </bk-form-item>
    <bk-form-item v-if="props.dualSystemSupport" :label="$t('客户端操作系统')" property="systemType">
      <bk-radio v-model="formData.systemType" label="Unix" @change="handleChangeSys(formData.systemType as string)">
        Linux
      </bk-radio>
      <bk-radio
        v-model="formData.systemType"
        label="Windows"
        @change="handleChangeSys(formData.systemType as string)" />
    </bk-form-item>
    <bk-form-item v-if="props.directoryShow" property="tempDir" :required="props.directoryShow">
      <template #label>
        {{ $t('临时目录') }}
        <info
          class="icon-info"
          v-bk-tooltips="{
            content: tempDirToolTips,
            placement: 'top',
          }" />
      </template>
      <bk-input v-model="formData.tempDir" :placeholder="$t('请输入')" clearable />
    </bk-form-item>
    <bk-form-item v-if="props.directoryShow">
      <div class="directory-description" :class="{ 'offset-margin': tempDirValidateStatus }">
        {{ t('客户端下载配置文件后，会将其保存在') }}
        <span
          v-bk-tooltips="{
            content: $t('一键复制'),
            placement: 'top',
          }"
          class="description-em"
          @click="handleCopyText(realPath)">
          &nbsp;{{ realPath }}&nbsp;
        </span>
      </div>
    </bk-form-item>
    <bk-form-item>
      <!-- 添加标签 -->
      <AddLabel ref="addLabelRef" @send-label="formData.labelArr = $event" />
    </bk-form-item>
    <!-- <bk-form-item v-if="p2pShow">
      由于集群列表接口暂不支持，产品将下拉框改为输入框，待后续接口支持后改回下拉框
      <p2p-acceleration
        ref="p2pAccelerationRef"
        @send-cluster="
          ({ clusterSwitch, clusterInfo }) => {
            formData.clusterSwitch = clusterSwitch;
            formData.clusterInfo = clusterInfo;
          }
        " />
    </bk-form-item> -->
    <bk-form-item v-if="p2pShow">
      <p2p-label @send-switcher="formData.clusterSwitch = $event" />
    </bk-form-item>
    <bk-form-item
      v-if="p2pShow && formData.clusterSwitch"
      class="cluster-form-item"
      property="clusterInfo"
      :required="formData.clusterSwitch">
      <bk-input v-model.trim="formData.clusterInfo" :placeholder="$t('请输入')" clearable />
    </bk-form-item>
    <!-- 启用配置文件筛选 -->
    <bk-form-item v-if="associateConfigShow">
      <associate-config @update-rules="formData.rules = $event" />
    </bk-form-item>
    <!-- 节点管理插件换行符选择 -->
    <bk-form-item v-if="lineBreakShow" class="line-break-item">
      <template #label>
        {{ $t('文本文件换行符：') }}
        <info
          class="icon-info"
          v-bk-tooltips="{
            content: $t(`客户端下载文件时可以选择将文件保存为Linux格式（使用LF换行符）或Windows格式（使用CRLF换行符）
服务端默认使用Linux换行符进行保存，如果选择将文件保存为Windows格式，可能导致客户端文件的MD5值与服务端的MD5值不一致`),
            placement: 'top',
          }" />
      </template>
      <bk-select
        class="line-break-selector"
        v-model="formData.selectedLineBreak"
        :filterable="false"
        :clearable="false">
        <bk-option
          v-for="item in lineBreakData"
          :key="item"
          :name="item === 'CRLF' ? $t('CRLF（Windows换行符 \\r\\n）') : $t('LF（Linux换行符 \\n）')"
          :id="item">
          {{ item === 'CRLF' ? $t('CRLF（Windows换行符 \\r\\n）') : $t('LF（Linux换行符 \\n）') }}
        </bk-option>
      </bk-select>
    </bk-form-item>
  </bk-form>
</template>

<script lang="ts" setup>
  import { onMounted, ref, Ref, watch, inject, computed } from 'vue';
  import { useRoute } from 'vue-router';
  import KeySelect from './key-selector.vue';
  import { Info } from 'bkui-vue/lib/icon';
  import AddLabel from './add-label.vue';
  // import p2pAcceleration from './p2p-acceleration.vue';
  import p2pLabel from './p2p-label.vue';
  import { IExampleFormData, newICredentialItem } from '../../../../../../types/client';
  import { useI18n } from 'vue-i18n';
  import { cloneDeep } from 'lodash';
  import { copyToClipBoard } from '../../../../../utils/index';
  import BkMessage from 'bkui-vue/lib/message';
  import associateConfig from './associate-config.vue';
  import configSelector from './config-selector.vue';

  const props = withDefaults(
    defineProps<{
      selectedKeyData: newICredentialItem['spec'] | null; // 记忆密钥的信息
      serviceType: string; // 服务类型
      templateName?: string;
      directoryShow?: boolean; // 临时目录(所有文件型)
      p2pShow?: boolean; // p2p网络加速（Sidecar容器）
      configShow?: boolean; // 配置项名称（Python SDK、http(s)接口调用）
      configLabel?: string; // 配置项label
      associateConfigShow?: boolean; // 配置文件筛选功能（所有文件型）
      dualSystemSupport?: boolean; // Linux与Windows双系统支持（节点管理插件与两种类型的cmd命令行工具）
      lineBreakShow?: boolean; // 换行符选项(节点管理插件)
    }>(),
    {
      directoryShow: true,
      p2pShow: false,
      configShow: false,
      configLabel: '配置项名称',
      associateConfigShow: false,
      dualSystemSupport: false,
      lineBreakShow: false,
    },
  );

  const emits = defineEmits(['update-option-data', 'selected-key-data']);

  const { t } = useI18n();
  const route = useRoute();
  const sysDirectories: string[] = ['/bin', '/boot', '/dev', '/lib', '/lib64', '/proc', '/run', '/sbin', '/sys'];
  const lineBreakData = ['LF', 'CRLF'];

  const basicInfo = inject<{ serviceName: Ref<string> }>('basicInfo');
  const addLabelRef = ref();
  const keySelectorRef = ref();
  const configSelectRef = ref();
  // const p2pAccelerationRef = ref();
  const formRef = ref();
  const formData = ref<IExampleFormData>({
    clientKey: '', // 客户端密钥
    privacyCredential: '', // 脱敏的密钥
    tempDir: '/data/bscp', // 临时目录
    configName: '', // 配置项
    labelArr: [], // 添加的标签
    clusterSwitch: false, // 集群开关
    clusterInfo: 'BCS-K8S-', // 集群ID
    rules: [], // 文件筛选规则
    systemType: 'Unix', // 系统类型
    selectedLineBreak: 'LF', // 换行符
    // clusterInfo: {
    //   name: '', // 集群名称
    //   value: '', // 集群id
    // },
  });
  const spaceId = ref(Number(route.params.spaceId));

  const rules = {
    clientKey: [
      {
        required: true,
        message: t('请先选择客户端密钥，替换下方示例代码后，再尝试复制示例'),
        validator: (value: string) => value.length,
        trigger: 'change',
      },
    ],
    tempDir: [
      {
        required: true,
        message: t('请输入路径地址，替换下方示例代码后，再尝试复制示例'),
        validator: (value: string) => value.length,
        trigger: 'change',
      },
      {
        required: true,
        message: t('禁止使用系统目录'),
        validator: (value: string) => !sysDirectories.some((dir) => value === dir || value.startsWith(`${dir}/`)),
        trigger: 'change',
      },
      {
        required: true,
        validator: (value: string) => {
          // Unix与Windows双路径判断
          if (formData.value.systemType === 'Windows') {
            // return /^[a-zA-Z]:\\(?:[^\\/:*?"<>|\r\n]+\\)*[^\\/:*?"<>|\r\n]+$/.test(formData.value.tempDir);
            // return /^[a-zA-Z]:\\(?:[^\\/:*?"<>|\r\n]+\\)*[^\\/:*?"<>|\r\n]+$|^[a-zA-Z]:$/.test(formData.value.tempDir);/^[A-Za-z]:\$/
            return /^(?:[A-Za-z]:\\$|[A-Za-z]:\\(?:[^\\/:*?"<>|\r\n]+\\)*[^\\/:*?"<>|\r\n]+$|^[a-zA-Z]:$)/.test(
              formData.value.tempDir,
            );
          }
          // 单Unix路径判断
          // 必须为绝对路径, 且不能以/结尾
          if (!value.startsWith('/') || value.endsWith('/')) {
            return false;
          }
          const parts = value.split('/').slice(1);
          let isValid = true;
          // 文件路径校验
          parts.some((part) => {
            if (part.startsWith('.') || !/^[\u4e00-\u9fa5A-Za-z0-9.\-_#%,@^+=\\[\]{}]+$/.test(part)) {
              isValid = false;
              return true;
            }
            return false;
          });
          return isValid;
        },
        trigger: 'change',
        message: () => {
          return t('无效的路径,路径不符合systemType文件路径格式规范', {
            systemType: formData.value.systemType,
          });
        },
      },
    ],
    configName: [
      {
        required: true,
        validator: (value: string) => value.length <= 128,
        message: t('最大长度128个字符'),
        trigger: 'change',
      },
      {
        required: true,
        validator: (value: string) => value.length,
        message: t('请先选择配置项名称，替换下方示例代码后，再尝试复制示例'),
        trigger: 'change',
      },
      // {
      //   required: true,
      //   validator: (value: string) =>
      //     /^[\p{Script=Han}\p{L}\p{N}]([\p{Script=Han}\p{L}\p{N}_-]*[\p{Script=Han}\p{L}\p{N}])?$/u.test(value),
      //   message: t('只允许包含中文、英文、数字、下划线 (_)、连字符 (-)，并且必须以中文、英文、数字开头和结尾'),
      //   trigger: 'change',
      // },
    ],
    clusterInfo: [
      {
        required: true,
        message: t('请输入BCS 集群 ID，替换下方示例代码后，再尝试复制示例'),
        validator: (value: string) => value.length,
        trigger: 'blur',
      },
      {
        required: true,
        validator: (value: string) => /^BCS-K8S-\d{5}$/.test(value),
        message: t('BCS集群ID须符合以下格式：BCS-K8S-xxxxx，其中xxxxx为5位数字'),
        trigger: 'blur',
      },
    ],
  };

  // 临时目录校验，用于底部提示文案摆放位置
  const tempDirValidateStatus = computed(() => {
    return rules.tempDir.every((ruleItem) => ruleItem.validator(formData.value.tempDir));
  });

  // 真实路径
  const realPath = computed(() => {
    if (formData.value.systemType === 'Windows') {
      return `${formData.value.tempDir}${/^[A-Za-z]:\\$/.test(formData.value.tempDir) ? '' : '\\'}${spaceId.value}\\${basicInfo?.serviceName.value}\\files`;
    }
    return `${formData.value.tempDir}/${spaceId.value}/${basicInfo?.serviceName.value}/files`;
  });

  const tempDirToolTips = computed(() => {
    if (formData.value.systemType === 'Windows') {
      return t('临时目录提示文案').replaceAll('/', '\\');
    }
    return t('临时目录提示文案');
  });

  watch(formData.value, () => {
    sendAll();
    // tempDirPathType(formData.value.tempDir);
  });

  onMounted(() => {
    sendAll();
  });

  // 选择操作系统改变默认路径
  const handleChangeSys = (type: string) => {
    formData.value.tempDir = type === 'Windows' ? 'D:\\bscp' : '/data/bscp';
    formData.value.selectedLineBreak = type === 'Windows' ? 'CRLF' : 'LF';
  };

  const handleValidate = () => {
    // label验证，数组长度为空时返回true
    const labelValid = addLabelRef.value.isAllValid();
    // p2p网络加速验证，目前只有Sidecar使用，根据有无使用决定验证情况
    // const p2pValid = props.p2pShow ? p2pAccelerationRef.value.isValid() : true;
    // 密钥验证
    const keyValid = keySelectorRef.value.validateCredential();
    const configValid = props.configShow ? configSelectRef.value.validateConfig() : true;
    // const isAllValid = [labelValid, p2pValid, keyValid].includes(false);
    const isAllValid = [labelValid, keyValid, configValid].includes(false);
    if (isAllValid) {
      formRef.value.validate();
      return Promise.reject();
    }
    return formRef.value.validate();
  };

  const setCredential = (key: string, privacyKey: string) => {
    formData.value.clientKey = key;
    formData.value.privacyCredential = privacyKey;
  };

  // 复制
  const handleCopyText = async (text: string) => {
    try {
      await formRef.value.validate('tempDir');
      copyToClipBoard(text);
      BkMessage({
        theme: 'success',
        message: t('目录复制成功'),
      });
    } catch (error) {
      BkMessage({
        theme: 'error',
        message: error,
      });
    }
  };

  const sendAll = () => {
    const filterFormData = cloneDeep(formData.value);
    // 临时目录不合法的路径不发送
    if (!tempDirValidateStatus.value) {
      filterFormData.tempDir = '';
    }
    emits('update-option-data', filterFormData);
  };

  defineExpose({
    handleValidate,
  });
</script>

<style scoped lang="scss">
  .headline {
    font-size: 14px;
    font-weight: 700;
    line-height: 22px;
    color: #63656e;
  }
  .form-example-wrap {
    margin-top: 16px;
    width: 537px;
    :deep(.bk-form-label) {
      font-size: 12px;
      & > span {
        position: relative;
      }
    }
  }
  .icon-info {
    position: absolute;
    right: -33px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 14px;
    color: #979ba5;
    cursor: pointer;
  }
  .cluster-form-item {
    margin-top: -18px;
  }
  .directory-description {
    margin: -6px 0 4px 0;
    display: flex;
    align-items: center;
    font-size: 12px;
    line-height: 18px;
    color: #979ba5;
    white-space: nowrap;
    overflow: hidden;
    transition: margin 0.1s;
    &.offset-margin {
      margin-top: -20px;
    }
    .copy-icon {
      margin-left: 12px;
      font-size: 14px;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
  }
  .description-em {
    margin-left: 4px;
    padding: 0 4px;
    flex: 0 1 auto;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    background-color: #f5f7fa;
    cursor: pointer;
    &:hover {
      color: #3a84ff;
      background-color: #f0f5ff;
    }
  }
</style>
