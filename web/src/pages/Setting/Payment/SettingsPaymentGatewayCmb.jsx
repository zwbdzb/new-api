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
import { Banner, Button, Form, Row, Col, Spin } from '@douyinfe/semi-ui';
import {
  API,
  showError,
  showSuccess,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';
import { BookOpen, TriangleAlert } from 'lucide-react';

export default function SettingsPaymentGatewayCmb(props) {
  const { t } = useTranslation();
  const sectionTitle = props.hideSectionTitle ? undefined : t('招商银行支付设置');
  const [loading, setLoading] = useState(false);
  // 只显示可编辑的非敏感配置项
  const [inputs, setInputs] = useState({
    CmbBaseURL: 'https://api.cmburl.cn:8065',
    CmbNotifyPath: '/api/user/zs_pay/notify',
    CmbPayValidTime: '1800',
    CmbMinTopUp: 1,
  });
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      // 从 options 中读取可编辑的配置项
      const currentInputs = {
        CmbBaseURL: props.options.CmbBaseURL || 'https://api.cmburl.cn:8065',
        CmbNotifyPath: props.options.CmbNotifyPath || '/api/user/zs_pay/notify',
        CmbPayValidTime: props.options.CmbPayValidTime || '1800',
        CmbMinTopUp:
          props.options.CmbMinTopUp !== undefined
            ? parseFloat(props.options.CmbMinTopUp)
            : 1,
      };
      setInputs(currentInputs);
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitCmbSetting = async () => {
    if (props.options.ServerAddress === '') {
      showError(t('请先填写服务器地址'));
      return;
    }

    setLoading(true);
    try {
      const options = [];

      // 只保存可编辑的非敏感配置项
      if (inputs.CmbBaseURL && inputs.CmbBaseURL !== '') {
        options.push({ key: 'CmbBaseURL', value: inputs.CmbBaseURL });
      }
      if (inputs.CmbNotifyPath && inputs.CmbNotifyPath !== '') {
        options.push({ key: 'CmbNotifyPath', value: inputs.CmbNotifyPath });
      }
      if (
        inputs.CmbPayValidTime !== undefined &&
        inputs.CmbPayValidTime !== null &&
        inputs.CmbPayValidTime !== ''
      ) {
        options.push({
          key: 'CmbPayValidTime',
          value: inputs.CmbPayValidTime.toString(),
        });
      }
      if (
        inputs.CmbMinTopUp !== undefined &&
        inputs.CmbMinTopUp !== null
      ) {
        options.push({
          key: 'CmbMinTopUp',
          value: inputs.CmbMinTopUp.toString(),
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
        <Form.Section text={sectionTitle}>
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='CmbBaseURL'
                label={t('API 基础地址')}
                placeholder={t('https://api.cmburl.cn:8065')}
                extraText={t('测试环境：https://api.cmburl.cn:8065，生产环境：https://api.cmbchina.com')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='CmbNotifyPath'
                label={t('回调路径')}
                placeholder={t('/api/user/zs_pay/notify')}
                extraText={t('支付回调处理路径')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='CmbPayValidTime'
                label={t('二维码有效期（秒）')}
                placeholder={t('1800')}
                extraText={t('二维码有效时间，默认 1800 秒')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='CmbMinTopUp'
                label={t('最低充值数量')}
                placeholder={t('1')}
                extraText={t('用户单次最少可充值的数量')}
              />
            </Col>
          </Row>
          <Button onClick={submitCmbSetting} style={{ marginTop: 16 }}>{t('更新招商银行支付设置')}</Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
