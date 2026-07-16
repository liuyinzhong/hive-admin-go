-- 修复医生管理第一阶段字典中文标签和备注。
-- 原因：通过 Windows PowerShell 文本管道执行 UTF-8 SQL 时，中文曾被转换为问号。
-- 请使用 MySQL SOURCE 命令直接读取本文件，不要通过 PowerShell 文本管道传输。

SET NAMES utf8mb4;

UPDATE `sys_dict`
SET `label` = '医生性别', `remark` = '医生档案性别字典', `update_date` = NOW()
WHERE `type` = 'MED_DOCTOR_GENDER' AND `pid` IS NULL AND `del_flag` = 0;

UPDATE `sys_dict`
SET `label` = CASE `value`
  WHEN '0' THEN '未知'
  WHEN '1' THEN '男'
  WHEN '2' THEN '女'
  ELSE `label`
END, `update_date` = NOW()
WHERE `type` = 'MED_DOCTOR_GENDER' AND `value` IN ('0', '1', '2') AND `del_flag` = 0;

UPDATE `sys_dict`
SET `label` = '医生职称', `remark` = '医生专业技术职称字典', `update_date` = NOW()
WHERE `type` = 'MED_DOCTOR_TITLE' AND `pid` IS NULL AND `del_flag` = 0;

UPDATE `sys_dict`
SET `label` = CASE `value`
  WHEN '1' THEN '住院医师'
  WHEN '2' THEN '主治医师'
  WHEN '3' THEN '副主任医师'
  WHEN '4' THEN '主任医师'
  WHEN '5' THEN '其他'
  ELSE `label`
END, `update_date` = NOW()
WHERE `type` = 'MED_DOCTOR_TITLE' AND `value` IN ('1', '2', '3', '4', '5') AND `del_flag` = 0;

UPDATE `sys_dict`
SET `label` = '医生用工类型', `remark` = '医生档案用工类型字典', `update_date` = NOW()
WHERE `type` = 'MED_EMPLOYMENT_TYPE' AND `pid` IS NULL AND `del_flag` = 0;

UPDATE `sys_dict`
SET `label` = CASE `value`
  WHEN '1' THEN '全职'
  WHEN '2' THEN '兼职'
  WHEN '3' THEN '外聘'
  WHEN '4' THEN '多点执业'
  ELSE `label`
END, `update_date` = NOW()
WHERE `type` = 'MED_EMPLOYMENT_TYPE' AND `value` IN ('1', '2', '3', '4') AND `del_flag` = 0;

