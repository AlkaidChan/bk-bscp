<template>
  <Teleport :disabled="!isOpenFullScreen" to="body">
    <div
      :class="{ fullscreen: isOpenFullScreen }"
      @mouseenter="isShowOperationBtn = true"
      @mouseleave="isShowOperationBtn = false">
      <Card :title="props.title" :height="360" style="margin-bottom: 16px">
        <template #operation>
          <OperationBtn
            v-show="isShowOperationBtn"
            :is-open-full-screen="isOpenFullScreen"
            @refresh="loadChartData"
            @toggle-full-screen="isOpenFullScreen = !isOpenFullScreen" />
        </template>
        <template #head-suffix>
          <bk-select
            v-model="selectTime"
            :class="['time-selector', { 'en-time-selector': locale === 'en' }]"
            :filterable="false"
            :clearable="false">
            <bk-option v-for="item in selectorTimeList" :id="item.value" :key="item.value" :name="item.label" />
          </bk-select>
        </template>
        <bk-loading class="loading-wrap" :loading="loading">
          <div v-if="!isDataEmpty" ref="canvasRef" class="canvas-wrap">
            <Tooltip ref="tooltipRef" :is-custom="true" @jump="jumpToSearch">
              <div class="tooltips-list">
                <div
                  v-for="item in customTooltip"
                  :key="item.x"
                  class="tooltips-item"
                  @click="handleClickTooltipsItem(item)">
                  <div class="tooltip-left">
                    <div class="marker" :style="{ background: item.color }"></div>
                    <div class="name">{{ item.name }}</div>
                  </div>
                  <div class="tooltip-right">
                    <div class="value">{{ item.value }}</div>
                    <Share class="icon" />
                  </div>
                </div>
              </div>
            </Tooltip>
          </div>
          <bk-exception v-else type="empty" scene="part" :description="$t('暂无数据')">
            <template #type>
              <span class="bk-bscp-icon icon-bar-chart exception-icon" />
            </template>
          </bk-exception>
        </bk-loading>
      </Card>
    </div>
  </Teleport>
</template>

<script lang="ts" setup>
  import { ref, onMounted, watch, computed } from 'vue';
  import Card from '../../components/card.vue';
  import OperationBtn from '../../components/operation-btn.vue';
  import { DualAxes } from '@antv/g2plot';
  import { getClientPullCountData } from '../../../../../../api/client';
  import { IPullCount, IClinetCommonQuery } from '../../../../../../../types/client';
  import Tooltip from '../../components/tooltip.vue';
  import useClientStore from '../../../../../../store/client';
  import { storeToRefs } from 'pinia';
  import { useRouter } from 'vue-router';
  import { useI18n } from 'vue-i18n';
  import { Share } from 'bkui-vue/lib/icon';

  const { t, locale } = useI18n();

  const router = useRouter();

  const clientStore = useClientStore();
  const { searchQuery } = storeToRefs(clientStore);

  const props = defineProps<{
    bkBizId: string;
    appId: number;
    title: string;
    isDuplicates: boolean;
  }>();

  let dualAxes: DualAxes | null;
  const canvasRef = ref<HTMLElement>();
  const tooltipRef = ref();
  const selectTime = ref(7);
  const selectorTimeList = [
    {
      value: 7,
      label: t('近 {n} 天', { n: 7 }),
    },
    {
      value: 15,
      label: t('近 {n} 天', { n: 15 }),
    },
    {
      value: 30,
      label: t('近 {n} 天', { n: 30 }),
    },
  ];
  const data = ref<IPullCount>({
    time: [],
    time_and_type: [],
  });
  const loading = ref(false);
  const isOpenFullScreen = ref(false);
  const jumpSearchTime = ref('');
  const isShowOperationBtn = ref(false);
  const customTooltip = ref();

  const isDataEmpty = computed(() => !data.value.time.some((item) => item.count > 0));

  watch([() => selectTime.value, () => props.appId], async () => {
    await loadChartData();
    if (!isDataEmpty.value) {
      if (dualAxes) {
        dualAxes.changeData([data.value.time, data.value.time_and_type]);
      } else {
        initChart();
      }
    }
  });

  watch(
    () => searchQuery.value,
    async () => {
      await loadChartData();
      if (!isDataEmpty.value) {
        if (dualAxes) {
          dualAxes.changeData([data.value.time, data.value.time_and_type]);
        } else {
          initChart();
        }
      }
    },
    { deep: true },
  );

  watch(
    () => data.value.time,
    () => {
      if (isDataEmpty.value && dualAxes) {
        dualAxes!.destroy();
        dualAxes = null;
      }
    },
  );

  onMounted(async () => {
    await loadChartData();
    if (!isDataEmpty.value) {
      initChart();
    }
  });

  const loadChartData = async () => {
    const params: IClinetCommonQuery = {
      search: searchQuery.value.search,
      pull_time: selectTime.value,
      last_heartbeat_time: searchQuery.value.last_heartbeat_time,
      is_duplicates: props.isDuplicates,
    };
    try {
      loading.value = true;
      const res = await getClientPullCountData(props.bkBizId, props.appId, params);
      data.value.time = res.time || [];
      data.value.time_and_type =
        res.time_and_type?.map((item: any) => {
          switch (item.type) {
            case 'sidecar':
              item.name = `SideCar ${t('客户端')}`;
              break;
            case 'sdk':
              item.name = `SDK ${t('客户端')}`;
              break;
            case 'agent':
              item.name = t('主机插件客户端');
              break;
            case 'command':
              item.name = `CLI ${t('客户端')}`;
              break;
          }
          return item;
        }) || [];
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const initChart = () => {
    dualAxes = new DualAxes(canvasRef.value!, {
      data: [data.value.time, data.value.time_and_type],
      xField: 'time',
      yField: ['count', 'value'],
      yAxis: {
        value: {
          grid: {
            line: {
              style: {
                stroke: '#979BA5',
                lineDash: [4, 5],
              },
            },
          },
          min: 0,
        },
        count: {
          grid: null,
          min: 0,
        },
      },
      geometryOptions: [
        {
          geometry: 'line',
          lineStyle: {
            lineWidth: 2,
          },
          color: '#2C2599',
          label: {
            position: 'top',
          },
          point: {
            shape: 'circle',
          },
        },
        {
          geometry: 'column',
          isGroup: true,
          seriesField: 'type',
          columnWidthRatio: 0.3,
          color: ['#3E96C2', '#61B2C2', '#85CCA8', '#B5E0AB'],
        },
      ],
      legend: {
        position: 'bottom',
        itemName: {
          formatter: (text) => {
            if (text === 'count') {
              return t('总量');
            }
            return text;
          },
        },
      },
      tooltip: {
        fields: ['value', 'count'],
        showTitle: true,
        title: 'time',
        container: tooltipRef.value?.getDom(),
        enterable: true,
        customItems: (originalItems: any[]) => {
          jumpSearchTime.value = originalItems[0].title.replace(/\//g, '-');
          originalItems.forEach((item) => {
            if (item.name === 'count') {
              item.name = t('总量');
            } else {
              item.name = item.data.name;
            }
          });
          originalItems.unshift(originalItems.pop());
          customTooltip.value = originalItems;
          return originalItems;
        },
      },
    });
    dualAxes!.render();
  };

  const jumpToSearch = () => {
    const routeData = router.resolve({
      name: 'client-search',
      params: { appId: props.appId, bizId: props.bkBizId },
      query: {
        pull_time: `${jumpSearchTime.value} 00:00:00 - ${jumpSearchTime.value} 23:59:59`,
        heartTime: searchQuery.value.last_heartbeat_time,
      },
    });
    window.open(routeData.href, '_blank');
  };

  const handleClickTooltipsItem = (item: any) => {
    if (item.name === '总量') return;
    const routeData = router.resolve({
      name: 'client-search',
      params: { appId: props.appId, bizId: props.bkBizId },
      query: {
        pull_time: `${jumpSearchTime.value} 00:00:00 - ${jumpSearchTime.value} 23:59:59`,
        heartTime: searchQuery.value.last_heartbeat_time,
        client_type: item.data.type,
      },
    });
    window.open(routeData.href, '_blank');
  };
</script>

<style scoped lang="scss">
  .loading-wrap {
    height: 100%;
  }
  .time-selector {
    margin-left: 16px;
    width: 88px;
    :deep(.bk-input) {
      height: 26px;
    }
    &.en-time-selector {
      width: 120px;
    }
  }
  .fullscreen {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 5000;
    .card {
      height: 100vh !important;
      :deep(.operation-btn) {
        top: 0 !important;
      }
    }
  }
  :deep(.bk-exception) {
    height: 100%;
    justify-content: center;
    transform: translateY(-20px);
  }
  :deep(.g2-tooltip) {
    .g2-tooltip-list-item {
      .g2-tooltip-marker {
        border-radius: initial !important;
      }
    }
    .g2-tooltip-list-item:nth-child(1) {
      .g2-tooltip-marker {
        height: 2px !important;
        transform: translatey(-3px);
      }
    }
  }
  .tooltips-list {
    margin-bottom: 12px;
    .tooltips-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      height: 27px;
      cursor: pointer;
      &:nth-child(1) {
        .marker {
          height: 2px !important;
        }
        .tooltip-right {
          margin-right: 20px;
        }
        .icon {
          display: none !important;
        }
      }
      &:hover {
        background: #f5f7fa;
      }
      .tooltip-left {
        display: flex;
        align-items: center;
      }
      .tooltip-right {
        @extend .tooltip-left;
        margin-left: 20px;
        .icon {
          color: #3a84ff;
          margin-left: 8px;
          font-size: 12px;
        }
      }
      .marker {
        width: 8px;
        height: 8px;
        margin-right: 8px;
      }
    }
  }
</style>
