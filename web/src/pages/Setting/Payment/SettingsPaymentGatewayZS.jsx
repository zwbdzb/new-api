/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useState, useRef } from 'react';
import {
  Banner,
  Button,
  Form,
  Row,
  Col,
  Typography,
  Spin,
} from '@douyinfe/semi-ui';
const { Text } = Typography;
import {
  API,
  showError,
  showSuccess,
  verifyJSON,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsPaymentGatewayZS(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    ZSPayEnabled: false,
    ZSPayNotifyPath: '/api/user/zs_pay/notify',
    ZSPayPayValidTime: '1800',
    TopupGroupRatio: '',
    AmountOptions: '',
    AmountDiscount: '',
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      // 美化 JSON 展示
      let TopupGroupRatio = props.options.TopupGroupRatio || '';
      try {
        if (TopupGroupRatio) {
          TopupGroupRatio = JSON.stringify(
            JSON.parse(TopupGroupRatio),
            null,
            2,
          );
        }
      } catch {}

      let AmountOptions = props.options.AmountOptions || '';
      try {
        if (AmountOptions) {
          AmountOptions = JSON.stringify(
            JSON.parse(AmountOptions),
            null,
            2,
          );
        }
      } catch {}

      let AmountDiscount = props.options.AmountDiscount || '';
      try {
        if (AmountDiscount) {
          AmountDiscount = JSON.stringify(
            JSON.parse(AmountDiscount),
            null,
            2,
          );
        }
      } catch {}

      const currentInputs = {
        ZSPayEnabled: props.options.ZSPayEnabled === 'true' || props.options.ZSPayEnabled === true,
        ZSPayNotifyPath: props.options.ZSPayNotifyPath || '/api/user/zs_pay/notify',
        ZSPayPayValidTime: props.options.ZSPayPayValidTime || '1800',
        TopupGroupRatio: TopupGroupRatio,
        AmountOptions: AmountOptions,
        AmountDiscount: AmountDiscount,
      };
      setInputs(currentInputs);
      setOriginInputs({ ...currentInputs });
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(prev => ({ ...prev, ...values }));
  };

  const submitZSSetting = async () => {
    // 从 formApi 获取最新的表单值
    let formValues = {};
    if (formApiRef.current) {
      formValues = formApiRef.current.getValues() || {};
    }
    
    // 合并 inputs 和 formValues，确保使用最新的值
    const finalInputs = { ...inputs, ...formValues };

    // 检查服务器地址（仅在非空时检查）
    const serverAddress = props.options?.ServerAddress || '';
    if (serverAddress && serverAddress.trim() === '') {
      showError(t('请先填写服务器地址'));
      return;
    }

    // 充值分组倍率验证（仅在值发生变化且非空时验证）
    const topupGroupRatio = finalInputs.TopupGroupRatio || inputs.TopupGroupRatio || '';
    if (originInputs['TopupGroupRatio'] !== topupGroupRatio) {
      if (topupGroupRatio && topupGroupRatio.trim() !== '' && !verifyJSON(topupGroupRatio)) {
        showError(t('充值分组倍率不是合法的 JSON 字符串'));
        return;
      }
    }

    // 自定义充值数量选项验证
    const amountOptions = finalInputs.AmountOptions || inputs.AmountOptions || '';
    if (originInputs['AmountOptions'] !== amountOptions) {
      if (amountOptions && amountOptions.trim() !== '' && !verifyJSON(amountOptions)) {
        showError(t('自定义充值数量选项不是合法的 JSON 数组'));
        return;
      }
    }

    // 充值金额折扣配置验证
    const amountDiscount = finalInputs.AmountDiscount || inputs.AmountDiscount || '';
    if (originInputs['AmountDiscount'] !== amountDiscount) {
      if (amountDiscount && amountDiscount.trim() !== '' && !verifyJSON(amountDiscount)) {
        showError(t('充值金额折扣配置不是合法的 JSON 对象'));
        return;
      }
    }

    setLoading(true);
    try {
      const options = [];

      // 启用开关
      options.push({
        key: 'zs_payment.Enabled',
        value: (finalInputs.ZSPayEnabled || inputs.ZSPayEnabled || false) ? 'true' : 'false',
      });

      // 回调路径
      const notifyPath = finalInputs.ZSPayNotifyPath || inputs.ZSPayNotifyPath || '';
      if (notifyPath && notifyPath !== '') {
        options.push({ key: 'zs_payment.NotifyPath', value: notifyPath });
      }

      // 支付有效期
      const payValidTime = finalInputs.ZSPayPayValidTime || inputs.ZSPayPayValidTime || '1800';
      if (payValidTime !== undefined && payValidTime !== null && payValidTime !== '') {
        options.push({
          key: 'zs_payment.PayValidTime',
          value: payValidTime.toString(),
        });
      }

      // 充值分组倍率
      if (originInputs['TopupGroupRatio'] !== topupGroupRatio) {
        options.push({ key: 'TopupGroupRatio', value: topupGroupRatio });
      }

      // 自定义充值数量选项
      if (originInputs['AmountOptions'] !== amountOptions) {
        options.push({
          key: 'payment_setting.amount_options',
          value: amountOptions,
        });
      }

      // 充值金额折扣配置
      if (originInputs['AmountDiscount'] !== amountDiscount) {
        options.push({
          key: 'payment_setting.amount_discount',
          value: amountDiscount,
        });
      }

      // 发送请求
      const requestQueue = options.map((opt) =>
        API.put('/api/option/', {
          key: opt.key,
          value: opt.value,
        }),
      );

      const results = await Promise.all(requestQueue);

      // 检查所有请求是否成功
      const errorResults = results.filter((res) => !res.data.success);
      if (errorResults.length > 0) {
        errorResults.forEach((res) => {
          showError(res.data.message);
        });
      } else {
        showSuccess(t('更新成功'));
        setOriginInputs({ ...finalInputs });
        props.refresh?.();
      }
    } catch (error) {
      showError(t('更新失败'));
    }
    setLoading(false);
  };

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={handleFormChange}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={t('招商银行聚合支付设置')}>
          <Text>
            {t('招商银行聚合支付支持微信、支付宝、银联等支付方式。')}
            <br />
          </Text>
          <Banner
            type='warning'
            description={t(
              '回调地址格式：服务器地址 + 回调路径，例如：https://your-domain.com/api/user/zs_pay/notify',
            )}
          />

          <Form.Switch
            field='ZSPayEnabled'
            label={t('启用招商银行聚合支付')}
            size='default'
            checkedText='｜'
            uncheckedText='〇'
            style={{ marginBottom: 16, display: 'block' }}
          />

          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='ZSPayNotifyPath'
                label={t('回调路径')}
                placeholder={t('/api/user/zs_pay/notify')}
                extraText={t('支付结果通知接收路径')}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.InputNumber
                field='ZSPayPayValidTime'
                label={t('支付有效期（秒）')}
                placeholder={t('1800')}
                min={60}
                step={60}
                extraText={t('二维码有效时间，默认 30 分钟（1800 秒）')}
              />
            </Col>
          </Row>

          <Form.TextArea
            field='TopupGroupRatio'
            label={t('充值分组倍率')}
            placeholder={t('为一个 JSON 文本，键为组名称，值为倍率')}
            autosize
            style={{ marginTop: 16 }}
          />

          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col span={24}>
              <Form.TextArea
                field='AmountOptions'
                label={t('自定义充值数量选项')}
                placeholder={t(
                  '为一个 JSON 数组，例如：[10, 20, 50, 100, 200, 500]',
                )}
                autosize
                extraText={t(
                  '设置用户可选择的充值数量选项，例如：[10, 20, 50, 100, 200, 500]',
                )}
              />
            </Col>
          </Row>

          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col span={24}>
              <Form.TextArea
                field='AmountDiscount'
                label={t('充值金额折扣配置')}
                placeholder={t(
                  '为一个 JSON 对象，例如：{"100": 0.95, "200": 0.9, "500": 0.85}',
                )}
                autosize
                extraText={t(
                  '设置不同充值金额对应的折扣，键为充值金额，值为折扣率，例如：{"100": 0.95, "200": 0.9, "500": 0.85}',
                )}
              />
            </Col>
          </Row>

          <Button onClick={submitZSSetting} style={{ marginTop: 16 }}>
            {t('更新招商银行聚合支付设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
