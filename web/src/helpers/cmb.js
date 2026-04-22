/**
 * Copyright (C) 2025 QuantumNous
 * 
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 * 
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 * 
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 * 
 * For commercial licensing, please contact support@quantumnous.com
 */

/**
 * 将招商银行二维码字符串转换为可扫描的二维码图片URL
 * @param {string} qrCode - 二维码内容字符串
 * @returns {string} 二维码图片URL
 */
export const QRCodeToURL = (qrCode) => {
  if (!qrCode) {
    return '';
  }
  // 如果已经是URL，直接返回
  if (qrCode.startsWith('http://') || qrCode.startsWith('https://')) {
    return qrCode;
  }
  // 使用第三方二维码生成服务生成二维码图片
  const encodedData = encodeURIComponent(qrCode);
  return `https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${encodedData}`;
};
