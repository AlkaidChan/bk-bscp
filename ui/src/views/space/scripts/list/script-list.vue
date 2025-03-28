<template>
  <section class="scripts-manage-page">
    <div class="side-menu">
      <div class="group-wrapper">
        <li :class="['group-item', { actived: showAllTag }]" @click="handleSelectTag('', true)">
          <i class="bk-bscp-icon icon-block-shape group-icon"></i>{{ t('全部脚本') }}
        </li>
        <li :class="['group-item', { actived: !showAllTag && selectedTag === '' }]" @click="handleSelectTag('')">
          <i class="bk-bscp-icon icon-tags group-icon"></i>{{ t('未分类') }}
        </li>
      </div>
      <div class="custom-tag-list">
        <li
          v-for="(item, index) in tagsData"
          :key="index"
          :class="['group-item', { actived: selectedTag === item.tag }]"
          @click="handleSelectTag(item.tag)">
          <i class="bk-bscp-icon icon-tags group-icon"></i>
          <span class="name">{{ item.tag }}</span>
          <span class="num">{{ item.counts }}</span>
        </li>
      </div>
    </div>
    <div class="script-list-wrapper">
      <div class="operate-area">
        <div class="btns">
          <bk-button theme="primary" @click="showCreateScript = true">
            <Plus class="button-icon" />
            {{ t('新建脚本') }}
          </bk-button>
          <BatchDeleteBtn
            :bk-biz-id="spaceId"
            :selected-ids="selectedIds"
            :is-across-checked="isAcrossChecked"
            :data-count="selecTableDataCount"
            @deleted="refreshAfterBatchDelete" />
        </div>
        <bk-input
          v-model="searchStr"
          class="search-script-input"
          :placeholder="t('脚本名称')"
          :clearable="true"
          @clear="refreshList()"
          @input="handleNameInputChange">
          <template #suffix>
            <Search class="search-input-icon" />
          </template>
        </bk-input>
      </div>
      <bk-loading style="min-height: 300px" :loading="scriptsLoading">
        <bk-table
          :border="['outer']"
          :data="scriptsData"
          :remote-pagination="true"
          :pagination="pagination"
          :class="memoEditHookId > 0 || tagEditHookId > 0 ? 'table-with-memo-edit' : ''"
          show-overflow-tooltip
          :cell-class="getCellCls"
          :row-class="getRowCls"
          @page-limit-change="handlePageLimitChange"
          @page-value-change="refreshList($event, true)">
          <template #prepend>
            <render-table-tip />
          </template>
          <bk-table-column :min-width="80" :width="80" :label="renderSelection" :show-overflow-tooltip="false">
            <template #default="{ row }">
              <across-check-box
                :checked="isChecked(row)"
                :disabled="row.bound_num !== 0"
                :handle-change="() => handleSelectionChange(row)" />
            </template>
          </bk-table-column>
          <bk-table-column :label="t('脚本名称')" :min-width="200">
            <template #default="{ row }">
              <div v-if="row.hook" class="hook-name" @click="handleViewVersionClick(row.hook.id)">
                {{ row.hook.spec.name }}
              </div>
            </template>
          </bk-table-column>
          <bk-table-column prop="hook.spec.type" :label="t('脚本语言')" :width="locale === 'zh-cn' ? '120' : '150'">
          </bk-table-column>
          <bk-table-column :label="t('分类标签')" property="tag">
            <template #default="{ row }">
              <div v-if="row.hook" class="script-tags">
                <div v-if="tagEditHookId !== row.hook.id" class="tags-display">
                  <ContentWidthOverflowTips>
                    <div v-if="row.hook.spec.tags?.length > 0" class="script-tags-list-wrapper">
                      <div v-for="tag in row.hook.spec.tags" class="tag-item" :key="tag">{{ tag }}</div>
                    </div>
                    <template v-else>--</template>
                  </ContentWidthOverflowTips>
                  <span class="edit-icon" @click="handleOpenTagEdit(row.hook.id)">
                    <EditLine />
                  </span>
                </div>
                <div v-else class="tag-edit-wrapper">
                  <bk-tag-input
                    :model-value="row.hook.spec.tags"
                    ref="tagInputRef"
                    display-key="tag"
                    save-key="tag"
                    search-key="tag"
                    :placeholder="t('请选择标签或输入新标签按Enter结束')"
                    :list="tagsData"
                    :allow-create="true"
                    trigger="focus"
                    @blur="
                      (inputVal: string, tagList: string[]) => {
                        handleTagEditBlur(row, inputVal, tagList);
                      }
                    " />
                </div>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('脚本描述')" property="memo">
            <template #default="{ row }">
              <div v-if="row.hook" class="script-memo">
                <div v-if="memoEditHookId !== row.hook.id" class="memo-display">
                  <bk-overflow-title class="memo-text" type="tips">
                    {{ row.hook.spec.memo || '--' }}
                  </bk-overflow-title>
                  <span class="edit-icon" @click="handleOpenMemoEdit(row)">
                    <EditLine />
                  </span>
                </div>
                <bk-input
                  v-else
                  ref="memoInputRef"
                  class="memo-input"
                  type="textarea"
                  v-model="editMemoStr"
                  :autosize="{ maxRows: 4 }"
                  :resize="false"
                  :maxlength="200"
                  @blur="handleMemoEditBlur(row, $event)" />
              </div>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('被引用')" width="100">
            <template #default="{ row }">
              <bk-button v-if="row.bound_num > 0" text theme="primary" @click="handleOpenCitedSlider(row.hook.id)">
                {{ row.bound_num }}
              </bk-button>
              <span v-else>0</span>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('更新人')" prop="hook.revision.reviser" width="140"></bk-table-column>
          <bk-table-column :label="t('更新时间')" width="180">
            <template #default="{ row }">
              <span v-if="row.hook">{{ datetimeFormat(row.hook.revision.update_at) }}</span>
            </template>
          </bk-table-column>
          <bk-table-column :label="t('操作')" :width="locale === 'zh-cn' ? '180' : '250'">
            <template #default="{ row }">
              <div class="action-btns">
                <bk-button text theme="primary" @click="handleEditClick(row)">{{ t('编辑') }}</bk-button>
                <bk-button
                  text
                  theme="primary"
                  @click="router.push({ name: 'script-version-manage', params: { spaceId, scriptId: row.hook.id } })">
                  {{ t('版本管理') }}
                </bk-button>
                <bk-button
                  :disabled="row.bound_num !== 0"
                  text
                  theme="primary"
                  v-bk-tooltips="{ content: $t('脚本已被引用，不能删除'), disabled: row.bound_num === 0 }"
                  @click="handleDeleteScript(row)">
                  {{ t('删除') }}
                </bk-button>
              </div>
            </template>
          </bk-table-column>
          <template #empty>
            <TableEmpty :is-search-empty="isSearchEmpty" @clear="clearSearchStr"></TableEmpty>
          </template>
        </bk-table>
      </bk-loading>
    </div>
    <CreateScript v-if="showCreateScript" v-model:show="showCreateScript" @created="handleCreatedScript" />
    <ScriptCited v-model:show="showCiteSlider" :id="currentId" />
  </section>
  <DeleteConfirmDialog
    v-model:is-show="isDeleteScriptDialogShow"
    :title="t('确认删除该脚本？')"
    @confirm="handleDeleteScriptConfirm">
    <div style="margin-bottom: 8px">
      {{ t('脚本') }}: <span style="color: #313238; font-weight: 600">{{ deleteScriptItem?.hook.spec.name }}</span>
    </div>
    <div style="margin-bottom: 8px">
      {{ t('一旦删除，该操作将无法撤销，以下服务配置的未命名版本中引用该脚本也将清除') }}
    </div>
    <div class="service-table">
      <bk-loading style="min-height: 200px" :loading="appsLoading">
        <bk-table :data="appList" :max-height="maxTableHeight" :empty-text="t('暂无未命名版本引用此脚本')">
          <bk-table-column :label="t('引用此脚本的服务')">
            <template #default="{ row }">
              <div class="app-info" @click="goToConfigPageImport(row.app_id)">
                <div v-overflow-title class="name-text">{{ row.app_name }}</div>
                <LinkToApp class="link-icon" :id="row.app_id" :auto-jump="true" />
              </div>
            </template>
          </bk-table-column>
        </bk-table>
      </bk-loading>
    </div>
  </DeleteConfirmDialog>
</template>
<script setup lang="ts">
  import { ref, watch, onMounted, computed, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import Message from 'bkui-vue/lib/message';
  import { Plus, Search, EditLine } from 'bkui-vue/lib/icon';
  import { storeToRefs } from 'pinia';
  import useGlobalStore from '../../../../store/global';
  import useScriptStore from '../../../../store/script';
  import {
    getScriptList,
    getScriptTagList,
    deleteScript,
    updateScript,
    getScriptCiteList,
  } from '../../../../api/script';
  import { IScriptItem, IScriptTagItem, IScriptListQuery } from '../../../../../types/script';
  import useTablePagination from '../../../../utils/hooks/use-table-pagination';
  import { datetimeFormat } from '../../../../utils/index';
  import BatchDeleteBtn from './batch-delete-btn.vue';
  import CreateScript from './create-script.vue';
  import ScriptCited from './script-cited.vue';
  import TableEmpty from '../../../../components/table/table-empty.vue';
  import DeleteConfirmDialog from '../../../../components/delete-confirm-dialog.vue';
  import LinkToApp from '../../templates/list/components/link-to-app.vue';
  import ContentWidthOverflowTips from '../../../../components/content-width-overflow-tips/index.vue';
  import useTableAcrossCheck from '../../../../utils/hooks/use-table-acrosscheck';
  import acrossCheckBox from '../../../../components/across-checkbox.vue';
  import CheckType from '../../../../../types/across-checked';
  import { debounce } from 'lodash';

  const { spaceId } = storeToRefs(useGlobalStore());
  const scriptStore = useScriptStore();
  const { versionListPageShouldOpenEdit, versionListPageShouldOpenView } = storeToRefs(scriptStore);
  const router = useRouter();
  const { t, locale } = useI18n();

  const { pagination, updatePagination } = useTablePagination('scriptCited');

  interface IAppItem {
    app_id: number;
    app_name: string;
  }
  const showCreateScript = ref(false);
  const showCiteSlider = ref(false);
  const scriptsData = ref<IScriptItem[]>([]);
  const scriptsLoading = ref(false);
  const tagsData = ref<IScriptTagItem[]>([]);
  const tagsLoading = ref(false);
  const showAllTag = ref(true); // 全部脚本
  const selectedTag = ref(''); // 未分类或具体tag下脚本
  const currentId = ref(0);
  const searchStr = ref('');
  const isDeleteScriptDialogShow = ref(false);
  const deleteScriptItem = ref<IScriptItem>();
  const appsLoading = ref(false);
  const appList = ref<IAppItem[]>([]);
  const selectedIds = ref<number[]>([]);
  const memoEditHookId = ref(0); // 当前正在编辑描述的脚本id
  const memoInputRef = ref();
  const tagEditHookId = ref(0); // 当前正在编辑标签的脚本id
  const tagInputRef = ref();
  const editMemoStr = ref(''); // 编辑描述内容
  const isAcrossChecked = ref(false);
  const selecTableDataCount = ref(0);
  const topIds = ref<number>();

  const maxTableHeight = computed(() => {
    const windowHeight = window.innerHeight;
    return windowHeight * 0.6 - 200;
  });

  // 当前页数据，不含禁用
  const selecTableData = computed(() => {
    return scriptsData.value
      .filter((item) => item.bound_num === 0)
      .map((item) => ({
        ...item,
        id: item.hook?.id,
      }));
  });

  const crossPageSelect = computed(
    () => pagination.value.limit < pagination.value.count && selecTableDataCount.value !== 0,
  );
  const { selectType, selections, renderSelection, renderTableTip, handleRowCheckChange, handleClearSelection } =
    useTableAcrossCheck({
      dataCount: selecTableDataCount,
      curPageData: selecTableData, // 当前页数据，不含禁用
      rowKey: ['id'],
      crossPageSelect,
    });

  const isSearchEmpty = ref(false);
  watch(
    () => spaceId.value,
    () => {
      refreshList();
      getTags();
    },
  );

  watch(
    selections,
    () => {
      isAcrossChecked.value = [CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(selectType.value);
      selectedIds.value = selections.value.map((item) => item.id);
    },
    {
      deep: true,
    },
  );

  onMounted(() => {
    getScripts();
    getTags();
  });

  // 获取脚本列表
  const getScripts = async (isCreate?: boolean) => {
    scriptsLoading.value = true;
    const params: IScriptListQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
    };
    if (selectedTag.value === '' && !showAllTag.value) {
      params.not_tag = true;
    } else if (selectedTag.value) {
      params.tag = selectedTag.value;
    }
    if (searchStr.value) {
      params.name = searchStr.value;
    }
    if (isCreate) {
      params.top_ids = topIds.value;
    } else {
      topIds.value = undefined;
    }

    const res = await getScriptList(spaceId.value, params);
    scriptsData.value = res.details;
    scriptsData.value.forEach((item) => {
      item.hook.spec.type = item.hook.spec.type.charAt(0).toUpperCase() + item.hook.spec.type.slice(1);
    });
    pagination.value.count = res.count;
    selecTableDataCount.value = Number(res.exclusion_count);
    scriptsLoading.value = false;
  };

  // 获取标签列表
  const getTags = async () => {
    tagsLoading.value = true;
    const res = await getScriptTagList(spaceId.value);
    tagsData.value = res.details;
    tagsLoading.value = false;
  };

  // 编辑脚本标签、描述
  const editScript = async (id: number, params: { memo: string; tags: string[] }) => {
    await updateScript(spaceId.value, id, params);
    Message({
      theme: 'success',
      message: t('脚本更新成功'),
    });
    getTags();
  };

  // 添加自定义单元格class
  const getCellCls = ({ property }: { property: string }) => {
    return ['tag', 'memo'].includes(property) ? 'memo-cell' : '';
  };

  // 表格行选择事件
  const handleSelectionChange = (row: IScriptItem) => {
    const isSelected = selections.value.some((item) => item.hook.id === row.hook.id);
    // 根据选择类型决定传递的状态
    const shouldBeChecked = isAcrossChecked.value ? isSelected : !isSelected;
    handleRowCheckChange(shouldBeChecked, { ...row, id: row.hook.id });
  };

  const handleSelectTag = (tag: string, all = false) => {
    searchStr.value = '';
    selectedTag.value = tag;
    showAllTag.value = all;
    refreshList();
  };

  // 触发编辑tag
  const handleOpenTagEdit = (id: number) => {
    tagEditHookId.value = id;
    nextTick(() => {
      tagInputRef.value?.focusInputTrigger();
    });
  };

  // 保存编辑后的tag
  const handleTagEditBlur = (script: IScriptItem, inputVal: string, tagList: string[]) => {
    tagEditHookId.value = 0;
    const { memo = '', tags = [] } = script.hook.spec;
    // 判断是否编辑过tag
    if (tagList.length !== tags.length || tagList.some((item) => !tags.includes(item))) {
      script.hook.spec.tags = tagList.slice();
      editScript(script.hook.id, { memo, tags: tagList });
    }
  };

  // 触发编辑脚本描述
  const handleOpenMemoEdit = (script: IScriptItem) => {
    const {
      id,
      spec: { memo },
    } = script.hook;
    memoEditHookId.value = id;
    editMemoStr.value = memo;
    nextTick(() => {
      memoInputRef.value?.focus();
    });
  };

  const handleMemoEditBlur = (script: IScriptItem, e: FocusEvent) => {
    memoEditHookId.value = 0;
    const { memo = '', tags = [] } = script.hook.spec;
    const val = (e.target as HTMLInputElement).value.trim();
    if (val !== memo) {
      script.hook.spec.memo = val;
      editScript(script.hook.id, { memo: val, tags });
    }
  };

  const handleOpenCitedSlider = (id: number) => {
    currentId.value = id;
    showCiteSlider.value = true;
  };

  const handleEditClick = (script: IScriptItem) => {
    router.push({ name: 'script-version-manage', params: { spaceId: spaceId.value, scriptId: script.hook.id } });
    versionListPageShouldOpenEdit.value = true;
  };

  const handleViewVersionClick = (id: number) => {
    router.push({ name: 'script-version-manage', params: { spaceId: spaceId.value, scriptId: id } });
    versionListPageShouldOpenView.value = true;
  };

  // 删除分组
  const handleDeleteScript = async (script: IScriptItem) => {
    const params = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
    };
    const res = await getScriptCiteList(spaceId.value, script.hook.id, params);
    const allAppInfo = res.details.map((item: any) => {
      const { app_id, app_name } = item;
      return {
        app_id,
        app_name,
      };
    });
    appList.value = Array.from(
      allAppInfo
        .reduce((map: Map<number, IAppItem>, obj: IAppItem) => {
          map.set(obj.app_id, obj);
          return map;
        }, new Map<number, IAppItem>())
        .values(),
    );
    deleteScriptItem.value = script;
    isDeleteScriptDialogShow.value = true;
  };

  const handleDeleteScriptConfirm = async () => {
    await deleteScript(spaceId.value, deleteScriptItem.value!.hook.id);
    if (scriptsData.value.length === 1 && pagination.value.current > 1) {
      pagination.value.current = pagination.value.current - 1;
    }
    Message({
      theme: 'success',
      message: t('删除脚本成功'),
    });
    isDeleteScriptDialogShow.value = false;
    getScripts();
  };

  const goToConfigPageImport = (id: number) => {
    const { href } = router.resolve({
      name: 'service-config',
      params: { appId: id },
    });
    window.open(href, '_blank');
  };

  const handleNameInputChange = debounce(() => refreshList(), 300);

  const handleCreatedScript = (topId: number) => {
    topIds.value = topId;
    refreshList(1, false, true);
    getTags();
  };

  const refreshAfterBatchDelete = () => {
    if (selectedIds.value.length === scriptsData.value.length && pagination.value.current > 1) {
      pagination.value.current -= 1;
    }

    selectedIds.value = [];
    refreshList(pagination.value.current);
  };

  const refreshList = (current = 1, pageChange = false, isCreate = false) => {
    // 非跨页全选/半选 需要重置全选状态
    if (![CheckType.HalfAcrossChecked, CheckType.AcrossChecked].includes(selectType.value) || !pageChange) {
      handleClearSelection();
    }
    isSearchEmpty.value = searchStr.value !== '';
    pagination.value.current = current;
    getScripts(isCreate);
  };

  const handlePageLimitChange = (val: number) => {
    updatePagination('limit', val);
    refreshList();
  };

  const clearSearchStr = () => {
    searchStr.value = '';
    refreshList();
  };

  const isChecked = (row: IScriptItem) => {
    if (![CheckType.AcrossChecked, CheckType.HalfAcrossChecked].includes(selectType.value)) {
      // 当前页状态传递
      return selections.value.some((item) => item.hook?.id === row.hook?.id);
    }
    // 跨页状态传递
    return !selections.value.some((item) => item.hook?.id === row.hook?.id);
  };

  const getRowCls = (script: IScriptItem) => {
    if (topIds.value === script.hook.id) {
      return 'new-row-marked';
    }
  };
</script>
<style lang="scss" scoped>
  .scripts-manage-page {
    display: flex;
    align-items: center;
    height: 100%;
    background: #ffffff;
  }
  .side-menu {
    padding: 16px 0;
    width: 280px;
    height: 100%;
    background: #f5f7fa;
    box-shadow: 0 2px 2px 0 rgba(0, 0, 0, 0.15);
    z-index: 1;
    .group-wrapper {
      padding-bottom: 16px;
      border-bottom: 1px solid #dcdee5;
    }
    .group-item {
      position: relative;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 8px 16px 8px 44px;
      color: #313238;
      font-size: 12px;
      cursor: pointer;
      &:hover {
        color: #348aff;
        .group-icon {
          color: #3a84ff;
        }
      }
      &.actived {
        background: #e1ecff;
        color: #3a84ff;
        .group-icon {
          color: #3a84ff;
        }
        .num {
          color: #ffffff;
          background: #a3c5fd;
        }
      }
      .group-icon {
        position: absolute;
        top: 9px;
        left: 22px;
        margin-right: 8px;
        font-size: 14px;
        color: #979ba5;
      }
      .name {
        flex: 0 1 auto;
        white-space: nowrap;
        text-overflow: ellipsis;
        overflow: hidden;
      }
      .num {
        flex: 0 0 auto;
        padding: 0 8px;
        height: 16px;
        line-height: 16px;
        color: #979ba5;
        background: #f0f1f5;
        border-radius: 2px;
      }
    }
    .custom-tag-list {
      padding: 16px 0;
      height: calc(100% - 82px);
      overflow: auto;
    }
  }
  .script-list-wrapper {
    padding: 24px;
    width: calc(100% - 280px);
    height: 100%;
    background: #ffffff;
    overflow: auto;
    :deep(.bk-table-body) {
      max-height: calc(100vh - 280px);
      overflow: auto;
      tr.new-row-marked td {
        background: #f2fff4 !important;
      }
    }
  }
  .operate-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    .btns {
      display: flex;
      align-items: center;
      :deep(.bk-button:not(:first-child)) {
        margin-left: 8px;
      }
    }
    .button-icon {
      font-size: 18px;
    }
  }
  .search-script-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    background: #ffffff;
  }
  .hook-name {
    color: #348aff;
    cursor: pointer;
  }
  .app-info {
    display: flex;
    align-items: center;
    overflow: hidden;
    cursor: pointer;
    .name-text {
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
    .link-icon {
      flex-shrink: 0;
      margin-left: 10px;
    }
  }
  .script-tags {
    position: relative;
    height: 100%;
    .tags-display {
      display: flex;
      align-items: center;
      height: 100%;
      overflow: hidden;
      &:hover {
        .edit-icon {
          display: block;
        }
      }
      .edit-icon {
        display: none;
        margin-left: 4px;
        font-size: 12px;
        color: #979ba5;
        cursor: pointer;
        &:hover {
          color: #3a84ff;
        }
      }
    }
    .tag-edit-wrapper {
      position: absolute;
      top: 4px;
      left: 0;
      right: 0;
      z-index: 2;
    }
  }
  .script-memo {
    .memo-display {
      display: flex;
      align-items: center;
      &:hover {
        .edit-icon {
          display: block;
        }
      }
    }
    .memo-text {
      margin-right: 4px;
      text-overflow: ellipsis;
      white-space: nowrap;
      overflow: hidden;
    }
    .edit-icon {
      display: none;
      font-size: 12px;
      color: #979ba5;
      cursor: pointer;
      &:hover {
        color: #3a84ff;
      }
    }
  }
  .action-btns {
    .bk-button {
      margin-right: 8px;
    }
  }
</style>

<style lang="scss">
  .service-table {
    thead th[colspan] {
      background-color: #f0f1f5 !important;
    }
  }
  .script-tags-list-wrapper {
    display: flex;
    align-items: center;
    overflow: hidden;
    .tag-item {
      flex-shrink: 0;
      padding: 0 10px;
      height: 22px;
      line-height: 22px;
      border-radius: 2px;
      font-size: 12px;
      background: #f0f1f5;
      color: #63656e;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      &:not(:last-child) {
        margin-right: 4px;
      }
    }
  }
</style>
