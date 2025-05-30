<template>
  <bk-form-item :label="t('配置项值')" property="value" required>
    <div :class="['secret-list', { disabled: isEdit }]">
      <div
        :class="['secret-item', { active: selectType === item.value }]"
        v-for="item in secretType"
        :key="item.value"
        @click="handleChangeCurrentType(item.value)">
        {{ item.label }}
      </div>
    </div>
    <div v-if="selectType === 'certificate' || selectType === 'custom'" class="certificate-content">
      <SecretContentEditor
        :content="secretValue"
        :is-credential="selectType === 'certificate'"
        :height="400"
        :is-edit="true"
        @change="handlSecretChange" />
      <div
        v-if="selectTypeContent!.infoList.length > 0"
        :class="['certificate-info', selectTypeContent?.infoList[0].status]">
        {{ selectTypeContent?.infoList[0].text }}
      </div>
    </div>
    <div v-else class="value-content">
      <bk-popover
        :is-show="isShowValidateInfo"
        ext-cls="secret-validate-info-popover"
        trigger="manual"
        placement="right"
        theme="light">
        <bk-input
          :model-value="isCipherShowSecret ? encryptValue : secretValue"
          class="value-input"
          :placeholder="inputPlaceholder"
          v-bk-tooltips="{
            content: inputPlaceholder,
            disabled: selectType !== 'token' || locale !== 'en',
          }"
          @input="handleValueChange"
          @focus="handleInputFocus"
          @blur="isShowValidateInfo = false">
          <template #suffix>
            <span
              v-if="selectTypeContent!.infoList.some((item) => item.status === 'warn') && secretValue"
              class="bk-bscp-icon icon-warning-circle warn-icon" />
            <Unvisible v-if="isCipherShowSecret" class="view-icon" @click="isCipherShowSecret = false" />
            <Eye v-else class="view-icon" @click="isCipherShowSecret = true" />
          </template>
        </bk-input>
        <template #content>
          <div class="info-list">
            <div v-for="item in selectTypeContent?.infoList" :key="item.text" class="info-item">
              <span :class="['dot', item.status]"></span> {{ item.text }}
            </div>
          </div>
        </template>
      </bk-popover>
    </div>
  </bk-form-item>
  <bk-checkbox
    class="visible-checkbox"
    :disabled="props.isEdit"
    :model-value="secretUnVisible"
    :before-change="handleChangeSecretUnVisible">
    {{ t('敏感信息不可见') }}
  </bk-checkbox>
  <bk-dialog
    :is-show="isShowVisibleDialog"
    :title="t('「敏感信息不可见」启用提示')"
    class="confirm-dialog"
    :width="480"
    @close="isShowVisibleDialog = false">
    <div class="dialog-content">
      <div v-for="(item, index) in enableTips" :key="index">
        {{ item }}
      </div>
    </div>
    <div class="dialog-footer">
      <bk-button theme="primary" @click="confirmEnable">{{ t('确定启用') }}</bk-button>
      <bk-button @click="isShowVisibleDialog = false">{{ t('取消') }}</bk-button>
    </div>
  </bk-dialog>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Eye, Unvisible } from 'bkui-vue/lib/icon';
  import { IConfigKvEditParams } from '../../../../../../../../../../types/config';
  import SecretContentEditor from './secret-content-editor.vue';
  import forge from 'node-forge';
  import { datetimeFormat } from '../../../../../../../../../utils';

  const { t, locale } = useI18n();

  const props = withDefaults(
    defineProps<{
      config: IConfigKvEditParams;
      isEdit: boolean;
    }>(),
    { isEdit: false },
  );

  const emits = defineEmits(['change']);

  const selectType = ref('password');

  const isCipherShowSecret = ref(true); // 密文展示敏感信息
  const secretUnVisible = ref(false); // 敏感信息不可见
  const isShowVisibleDialog = ref(false);
  const secretValue = ref('');
  const isShowValidateInfo = ref(false);
  const enableTips = [
    t('启用后，可降低数据泄露风险。尽管客户端拉去配置不受影响，但仍存在以下注意事项：'),
    t('⒈ 用户无法直接查看敏感数据，将导致查看和对比操作不便'),
    t('⒉ 编辑敏感信息时，需重新输入完整数据 '),
    t('⒊ 若忘记或丢失敏感信息，可能需要通过其他途径找回或重置'),
  ];

  onMounted(() => {
    const { secret_type, secret_hidden, value } = props.config;
    secretValue.value = secret_hidden ? '' : value;
    selectType.value = secret_type || 'password';
    secretUnVisible.value = secret_hidden;
  });

  const selectTypeContent = computed(() => {
    return secretType.find((item) => item.value === selectType.value);
  });

  const encryptValue = computed(() => {
    return '*'.repeat(secretValue.value.length);
  });

  const inputPlaceholder = computed(() => {
    return secretUnVisible.value && props.isEdit
      ? t('敏感数据不可见，无法查看实际内容；如需修改，请重新输入')
      : selectTypeContent.value?.placeholder;
  });

  const secretType = [
    {
      label: t('密码'),
      value: 'password',
      infoList: [
        { status: 'warn', text: t('建议长度 {n} 字符', { n: '8~64' }) },
        { status: 'warn', text: t('至少包含大写字母、小写字母、数字和特殊字符中的 3 种类型') },
      ],
      placeholder: t('请输入密码'),
    },
    {
      label: t('证书和私钥'),
      value: 'certificate',
      infoList: [],
    },
    {
      label: t('API密钥'),
      value: 'secret_key',
      infoList: [
        { status: 'warn', text: t('建议长度 {n} 字符', { n: '16~64' }) },
        { status: 'warn', text: t('包含大写字母、小写字母和数字') },
      ],
      placeholder: t('请输入 API 密钥'),
    },
    {
      label: t('访问令牌'),
      value: 'token',
      infoList: [
        { status: 'warn', text: t('建议长度 {n} 字符', { n: '32~512' }) },
        { status: 'warn', text: t('包含大写字母、小写字母和数字') },
      ],
      placeholder: t('请输入访问令牌，目前只支持 OAuth2.0 与 JWT 类型的访问令牌'),
    },
    {
      label: t('自定义'),
      value: 'custom',
      infoList: [],
    },
  ];

  const handleChangeCurrentType = (type: string) => {
    if (props.isEdit) return;
    secretValue.value = '';
    selectType.value = type;
    if (selectType.value === 'certificate') {
      selectTypeContent.value!.infoList = [];
    }
    change();
  };

  const handleValueChange = (value: string, event: any) => {
    if (!isCipherShowSecret.value) {
      // 明文展示 内容直接替换
      secretValue.value = value;
    } else {
      // 密文展示 内容处理
      // 输入的内容
      const editContent = value.replace(/\*/g, '');
      if (editContent) {
        // 添加或者修改内容 中间添加修改需要特殊处理
        // 正则表达式匹配前面的星号和对应的内容
        const startMatch = value.match(/^(\*)*/);
        const startAsterisks = startMatch ? startMatch[0].length : 0;
        const startContent = secretValue.value.slice(0, startAsterisks);
        // 正则表达式匹配后面的星号和对应的内容
        const endMatch = value.match(/(\*)*$/);
        const endAsterisks = endMatch ? endMatch[0].length : 0;
        const endContent = endAsterisks ? secretValue.value.slice(-endAsterisks) : '';
        secretValue.value = startContent + editContent + endContent;
      } else {
        // 删除的内容长度
        const deleteLength = secretValue.value.length - value.length;
        // 删除索引
        const deleteIndex = event.target.selectionStart;
        secretValue.value =
          secretValue.value.slice(0, deleteIndex) + secretValue.value.slice(deleteIndex + deleteLength);
      }
    }
    validateSecretValue();
    change();
  };

  const handleChangeSecretUnVisible = (val: boolean) => {
    if (val) {
      isShowVisibleDialog.value = true;
      change();
      return false;
    }
    secretUnVisible.value = false;
    return true;
  };

  const confirmEnable = () => {
    secretUnVisible.value = true;
    isShowVisibleDialog.value = false;
    change();
  };

  const handleInputFocus = () => {
    isShowValidateInfo.value = true;
    validateSecretValue();
  };

  // 更新状态的函数
  const updateStatus = (index: number, status: 'success' | 'warn') => {
    selectTypeContent.value!.infoList[index].status = status;
  };

  // 校验密钥内容
  const validateSecretValue = () => {
    const lengthConstraints = {
      password: { min: 8, max: 64 },
      secret_key: { min: 16, max: 64 },
      token: { min: 32, max: 512 },
    };

    // 类型断言
    const { min, max } = lengthConstraints[selectType.value as keyof typeof lengthConstraints] || {
      min: 0,
      max: Infinity,
    };
    const isValidLength = secretValue.value.length >= min && secretValue.value.length <= max;
    updateStatus(0, isValidLength ? 'success' : 'warn');

    const formatIsValid = checkSecretFormat(selectType.value === 'password');
    updateStatus(1, formatIsValid ? 'success' : 'warn');
  };

  // 判断input框内容格式
  const checkSecretFormat = (isPassword = true) => {
    if (secretValue.value.length < 3) return false;
    let hasUppercase = false; // 是否包含大写字母
    let hasLowercase = false; // 是否包含小写字母
    let hasDigit = false; // 是否包含数字
    let hasSpecialChar = false; // 是否包含特殊字符
    // 将字符串转成字符数组判断格式组成
    Array.from(secretValue.value).forEach((char) => {
      if (/[A-Z]/.test(char)) {
        hasUppercase = true;
      } else if (/[a-z]/.test(char)) {
        hasLowercase = true;
      } else if (/\d/.test(char)) {
        hasDigit = true;
      } else if (/[^A-Za-z\d]/.test(char)) {
        hasSpecialChar = true;
      }
    });
    if (isPassword) {
      // 计算满足的类型数量是否达到3个
      return [hasUppercase, hasLowercase, hasDigit, hasSpecialChar].filter(Boolean).length >= 3;
    }
    return hasUppercase && hasLowercase && hasDigit;
  };

  const handlSecretChange = (val: string) => {
    if (selectType.value === 'certificate') {
      validateCertificate(val);
      secretValue.value = val;
    } else {
      secretValue.value = val;
    }
    change();
  };

  const change = () => {
    emits('change', { value: secretValue.value, secret_type: selectType.value, visible: secretUnVisible.value });
  };

  const validateCertificate = (pem: string) => {
    if (!pem) {
      selectTypeContent.value!.infoList = [];
      return;
    }

    const results: Array<{ status: string; text: string }> = [];
    let hasError = false;
    let hasWarning = false;

    // 检测输入类型（证书或 RSA 私钥）
    const isCertificate = pem.includes('-----BEGIN CERTIFICATE-----');
    const isRsaPrivateKey = pem.includes('-----BEGIN RSA PRIVATE KEY-----');

    // 如果是 RSA 私钥（PKCS#1 格式）
    if (isRsaPrivateKey) {
      try {
        const pemCleaned = pem.replace(/-----(BEGIN|END) RSA PRIVATE KEY-----/g, '').replace(/[\r\n\t\s]/g, '');

        const der = forge.util.decode64(pemCleaned);
        forge.pki.privateKeyFromAsn1(forge.asn1.fromDer(der));

        results.push({
          status: 'normal',
        });
      } catch (e) {
        results.push({
          status: 'error',
          text: t('RSA 私钥格式无效'),
        });
      }
      selectTypeContent.value!.infoList = results;
      return;
    }

    // 如果不是证书格式，直接报错
    if (!isCertificate) {
      selectTypeContent.value!.infoList = [
        {
          status: 'error',
          text: t('输入格式不支持，请上传 X.509 证书或 RSA 私钥'),
        },
      ];
      return;
    }

    // 以下是证书校验逻辑
    const certPems = pem
      .split(/-----END CERTIFICATE-----\s*/)
      .filter((part) => part.trim().length > 0)
      .map((part) => `${part}-----END CERTIFICATE-----`);

    certPems.forEach((certPem) => {
      try {
        const pemCleaned = certPem.replace(/-----(BEGIN|END) CERTIFICATE-----/g, '').replace(/[\r\n\t\s]/g, '');

        const der = forge.util.decode64(pemCleaned);
        const cert = forge.pki.certificateFromAsn1(forge.asn1.fromDer(der));
        const notAfter = cert.validity.notAfter;
        const remainingDays = Math.ceil((notAfter.getTime() - Date.now()) / 86400000);

        if (remainingDays <= 0) {
          hasError = true;
          results.push({
            status: 'error',
            text: t('此证书已于 {n} 过期，请更换其它证书', { n: datetimeFormat(notAfter) }),
          });
        } else if (remainingDays <= 30) {
          hasWarning = true;
          results.push({
            status: 'warn',
            text: t('此证书将于 {n} 到期，距离到期仅剩 {m} 天', {
              n: datetimeFormat(notAfter),
              m: remainingDays,
            }),
          });
        } else {
          results.push({
            status: 'normal',
            text: t('此证书将于 {n} 到期，距离到期仅剩 {m} 天', { n: datetimeFormat(notAfter), m: remainingDays }),
          });
        }
      } catch (e) {
        hasError = true;
        results.push({
          status: 'error',
          text: t('证书解析失败：格式错误'),
        });
      }
    });

    // 结果优先级：错误 > 警告 > 正常
    if (hasError) {
      selectTypeContent.value!.infoList = results.filter((r) => r.status === 'error');
    } else if (hasWarning) {
      selectTypeContent.value!.infoList = results.filter((r) => r.status === 'warn');
    } else {
      // 默认显示第一个结果（如果没有错误或警告）
      selectTypeContent.value!.infoList = [results[0]];
    }
  };

  defineExpose({
    validate: () => {
      return !selectTypeContent.value?.infoList.some((info) => info.status === 'error');
    },
  });
</script>

<style scoped lang="scss">
  .secret-list {
    display: flex;
    align-items: center;
    margin-bottom: 12px;
    &.disabled {
      .secret-item {
        cursor: not-allowed;
        background: #f0f1f5;
        color: #979ba5;
      }
    }
    .secret-item {
      padding: 0 10px;
      height: 26px;
      min-width: 80px;
      line-height: 26px;
      background: #f8f8f8;
      text-align: center;
      background: #ffffff;
      border: 1px solid #c4c6cc;
      color: #63656e;
      cursor: pointer;
      &:not(:last-child) {
        border-right: none;
      }
      &.active {
        background: #e1ecff;
        border: 1px solid #3a84ff;
        color: #3a84ff;
      }
    }
  }
  .value-input {
    width: 560px;
    .view-icon {
      cursor: pointer;
      margin: 0 8px;
      font-size: 14px;
      color: #979ba5;
      &:hover {
        color: #3a84ff;
      }
    }
    .warn-icon {
      font-size: 14px;
      color: #ff9c01;
      line-height: 30px;
    }
  }

  .confirm-dialog {
    :deep(.bk-modal-wrapper) {
      .bk-modal-body {
        padding-bottom: 24px;
      }
      .bk-modal-header {
        .bk-dialog-title {
          text-align: center !important;
          white-space: initial;
        }
      }
      .dialog-content {
        padding: 12px 16px;
        width: 416px;
        background: #f5f6fa;
        border-radius: 2px;
        font-size: 14px;
        color: #63656e;
        line-height: 22px;
      }
      .dialog-footer {
        display: flex;
        justify-content: center;
        gap: 8px;
        margin-top: 24px;
      }
      .bk-modal-footer {
        display: none;
      }
    }
  }

  .secret-validate-info-popover {
    .info-list {
      .info-item {
        position: relative;
        padding-left: 12px;
        font-size: 12px;
        color: #63656e;
        line-height: 16px;
        &:not(:last-child) {
          margin-bottom: 8px;
        }
        .dot {
          position: absolute;
          display: inline-block;
          width: 6px;
          height: 6px;
          border-radius: 50%;
          left: 0;
          top: 5px;
          &.warn {
            background: #ff9c01;
          }
          &.success {
            background: #2dcb56;
          }
        }
      }
    }
  }

  .certificate-info {
    position: absolute;
    font-size: 12px;
    line-height: 20px;
    padding: 4px 0;
    &.normal {
      color: #c4c6cc;
    }
    &.warn {
      color: #ff9c01;
    }
    &.error {
      color: #ea3636;
    }
  }
</style>

<style>
  .secret-validate-info-popover.bk-popover.bk-pop2-content {
    max-width: 240px;
    padding: 8px;
  }
</style>
