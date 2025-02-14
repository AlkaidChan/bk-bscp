<template>
  <section class="config-list-with-templates">
    <SearchInput
      v-model="searchStr"
      class="config-search-input"
      :placeholder="t('配置文件名/创建人/修改人')"
      @search="getListData" />
    <bk-loading class="loading-wrapper" :loading="loading">
      <div v-for="group in tableGroupsData" :key="group.id" class="config-group">
        <template v-if="group.configs.length > 0">
          <div class="group-title" @click="group.expand = !group.expand">
            <DownShape :class="['fold-icon', { fold: !group.expand }]" />
            {{ group.name }}
          </div>
          <div v-if="group.expand" class="config-list-wrapper">
            <div
              v-for="config in group.configs"
              :class="['config-item', { disabled: config.file_state === 'DELETE' }]"
              :key="config.id"
              @click="handleConfigClick(config, group)">
              <bk-overflow-title class="config-name" type="tips">
                {{ fileAP(config) }}
              </bk-overflow-title>
              <div class="config-type">{{ getConfigTypeName(config.file_type) }}</div>
            </div>
          </div>
        </template>
      </div>
      <TableEmpty v-if="isTableEmpty" :is-search-empty="isSearchEmpty" @clear="clearSearch" />
    </bk-loading>
    <EditConfig
      v-model:show="editConfigSliderData.open"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      :config-id="editConfigSliderData.id" />
    <ViewConfig
      v-model:show="viewConfigSliderData.open"
      v-bind="viewConfigSliderData.data"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      :version-id="versionData.id"
      @open-edit="handleOpenEdit" />
  </section>
</template>
<script setup lang="ts">
  import { ref, watch, computed, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { storeToRefs } from 'pinia';
  import { DownShape } from 'bkui-vue/lib/icon';
  import useConfigStore from '../../../../../../../store/config';
  import { IConfigItem, IBoundTemplateGroup } from '../../../../../../../../types/config';
  import { ICommonQuery } from '../../../../../../../../types/index';
  import {
    getConfigList,
    getReleasedConfigList,
    getBoundTemplates,
    getBoundTemplatesByAppVersion,
  } from '../../../../../../../api/config';
  import { getConfigTypeName } from '../../../../../../../utils/config';
  import SearchInput from '../../../../../../../components/search-input.vue';
  import EditConfig from '../config-table-list/edit-config.vue';
  import ViewConfig from '../config-table-list/view-config.vue';
  import TableEmpty from '../../../../../../../components/table/table-empty.vue';
  import { debounce } from 'lodash';

  interface IConfigsGroupData {
    id: number;
    name: string;
    expand: boolean;
    configs: IConfigTableItem[];
    template_space_id?: number;
    template_space_name?: string;
    template_set_id?: number;
    template_set_name?: string;
  }

  interface IConfigTableItem {
    id: number;
    name: string;
    file_type: string;
    versionId: number;
    versionName: string;
    path: string;
    creator: string;
    reviser: string;
    update_at: string;
    file_state: string;
    is_latest?: boolean;
  }

  interface ITemplateConfigMeta {
    template_space_id: number;
    template_space_name: string;
    template_set_id: number;
    template_set_name: string;
  }

  const store = useConfigStore();
  const { versionData } = storeToRefs(store);
  const { t } = useI18n();

  const props = defineProps<{
    bkBizId: string;
    appId: number;
  }>();

  const loading = ref(false);
  const commonConfigListLoading = ref(false);
  const configList = ref<IConfigItem[]>([]); // 非模板配置文件
  const boundTemplateListLoading = ref(false);
  const templateGroupList = ref<IBoundTemplateGroup[]>([]); // 配置文件模板
  const tableGroupsData = ref<IConfigsGroupData[]>([]);
  const searchStr = ref('');
  const isSearchEmpty = ref(false);
  const editConfigSliderData = ref({
    open: false,
    id: 0,
  });
  const viewConfigSliderData = ref<{
    open: boolean;
    data: {
      id: number;
      type: string;
      templateMeta?: ITemplateConfigMeta;
      versionName?: string;
      isLatest?: boolean;
    };
  }>({
    open: false,
    data: {
      id: 0,
      type: '',
    },
  });

  watch(
    () => versionData.value.id,
    () => {
      getListData();
    },
  );

  watch(
    () => searchStr.value,
    (val) => {
      isSearchEmpty.value = !!val;
    },
  );

  // 是否为未命名版本
  const isUnNamedVersion = computed(() => versionData.value.id === 0);

  const isTableEmpty = computed(() => tableGroupsData.value.every((group) => group.configs.length === 0));

  onMounted(() => {
    getListData();
  });

  // 配置文件名
  const fileAP = (config: IConfigTableItem) => {
    const { path, name } = config;
    if (path.endsWith('/')) {
      return `${path}${name}`;
    }
    return `${path}/${name}`;
  };

  const getListData = debounce(async () => {
    const currentSearchStr = searchStr.value;
    // 拉取到版本列表之前不加在列表数据
    if (typeof versionData.value.id !== 'number') {
      return;
    }
    loading.value = true;
    await Promise.all([getCommonConfigList(), getBoundTemplateList()]);
    loading.value = false;
    // 处理文件数量过多 导致上一次搜索结果返回比这一次慢 导入搜索结果错误 取消数据处理
    if (currentSearchStr !== searchStr.value) return;
    tableGroupsData.value = transListToTableData();
  }, 500);

  // 获取非模板配置文件列表
  const getCommonConfigList = async () => {
    commonConfigListLoading.value = true;
    try {
      const params: ICommonQuery = {
        start: 0,
        all: true,
      };
      if (searchStr.value) {
        params.search_fields = 'name,path,memo,creator';
        params.search_value = searchStr.value;
      }
      let res;
      if (isUnNamedVersion.value) {
        res = await getConfigList(props.bkBizId, props.appId, params);
      } else {
        res = await getReleasedConfigList(props.bkBizId, props.appId, versionData.value.id, params);
      }

      configList.value = res.details;
    } catch (e) {
      console.error(e);
    } finally {
      commonConfigListLoading.value = false;
    }
  };

  // 获取模板配置文件列表
  const getBoundTemplateList = async () => {
    boundTemplateListLoading.value = true;
    try {
      const params: ICommonQuery = {
        start: 0,
        all: true,
      };
      if (searchStr.value) {
        params.search_fields = 'revision_name,revision_memo,name,path,creator';
        params.search_value = searchStr.value;
      }

      let res;
      if (isUnNamedVersion.value) {
        res = await getBoundTemplates(props.bkBizId, props.appId, params);
      } else {
        res = await getBoundTemplatesByAppVersion(props.bkBizId, props.appId, versionData.value.id, params);
      }
      templateGroupList.value = res.details;
    } catch (e) {
      console.error(e);
    } finally {
      boundTemplateListLoading.value = false;
    }
  };

  const transListToTableData = () => {
    const pkgsGroups = groupTplsByPkg(templateGroupList.value);
    return [
      { id: 0, name: t('非模板配置'), expand: true, configs: transConfigsToTableItemData(configList.value) },
      ...pkgsGroups,
    ];
  };

  // 将非模板配置文件数据转为表格数据
  const transConfigsToTableItemData = (list: IConfigItem[]) =>
    list.map((item: IConfigItem) => {
      const { id, spec, revision, file_state } = item;
      const { name, file_type, path } = spec;
      const { creator, reviser, update_at } = revision;
      return { id, name, versionId: 0, versionName: '--', path, creator, reviser, update_at, file_type, file_state };
    });

  // 将模板按套餐分组，并将模板数据格式转为表格数据
  const groupTplsByPkg = (list: IBoundTemplateGroup[]) => {
    const groups: IConfigsGroupData[] = list.map((groupItem) => {
      const { template_space_name, template_set_id, template_set_name, template_revisions, template_space_id } =
        groupItem;
      const group: IConfigsGroupData = {
        id: template_set_id,
        template_set_id,
        template_set_name,
        template_space_name,
        template_space_id,
        name: `${template_space_name} - ${template_set_name}`,
        expand: false,
        configs: [],
      };
      template_revisions.forEach((tpl) => {
        const {
          template_id: id,
          name,
          template_revision_id: versionId,
          template_revision_name: versionName,
          path,
          file_type,
          creator,
          file_state,
        } = tpl;
        group.configs.push({
          id,
          name,
          versionId,
          versionName,
          path,
          file_type,
          creator,
          reviser: '--',
          update_at: '--',
          file_state,
        });
      });
      return group;
    });
    return groups;
  };

  const handleConfigClick = (config: IConfigTableItem, group: IConfigsGroupData) => {
    const id = group.id === 0 ? config.id : config.versionId;
    if (group.id === 0) {
      viewConfigSliderData.value = {
        open: true,
        data: { id, type: 'config' },
      };
    } else {
      const { versionName, is_latest, id, versionId } = config;
      const { template_set_id, template_space_id, template_set_name, template_space_name } = group;
      const templateMeta = { template_space_id, template_space_name, template_set_id, template_set_name };
      const viewTemplateId = isUnNamedVersion.value ? id : versionId;
      viewConfigSliderData.value = {
        open: true,
        data: {
          id: viewTemplateId,
          versionName,
          templateMeta: templateMeta as ITemplateConfigMeta,
          type: 'template',
          isLatest: is_latest,
        },
      };
    }
    editConfigSliderData.value.id = id;
  };

  const handleOpenEdit = () => {
    viewConfigSliderData.value.open = false;
    editConfigSliderData.value.open = true;
  };

  const clearSearch = () => {
    searchStr.value = '';
    getListData();
  };
</script>
<style lang="scss" scoped>
  .config-list-with-templates {
    padding: 24px;
    height: 100%;
    background: #fafbfd;
    overflow: auto;
  }
  .config-search-input {
    margin-bottom: 16px;
  }
  .loading-wrapper {
    height: calc(100% - 48px);
    overflow: auto;
  }
  .group-title {
    display: flex;
    align-items: center;
    margin: 8px 0;
    line-height: 20px;
    font-size: 12px;
    color: #63656e;
    cursor: pointer;
    .fold-icon {
      margin-right: 8px;
      font-size: 14px;
      color: #3a84ff;
      transition: transform 0.2s ease-in-out;
      &.fold {
        color: #c4c6cc;
        transform: rotate(-90deg);
      }
    }
  }
  .config-list-wrapper {
    max-height: 472px; // 每个分组最多显示10条，超出后滚动显示
    overflow: auto;
  }
  .config-item {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
    font-size: 12px;
    background: #ffffff;
    box-shadow: 0 1px 1px 0 rgba(0, 0, 0, 0.06);
    border-radius: 2px;
    cursor: pointer;
    &.disabled {
      cursor: not-allowed;
      .config-type,
      .config-name {
        color: #dcdee5;
      }
    }
    &:not(.disabled):hover {
      background: #e1ecff;
    }
    .config-name {
      padding: 0 16px;
      width: 242px;
      height: 40px;
      line-height: 40px;
      color: #313238;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .config-type {
      color: #979ba5;
    }
  }
</style>
